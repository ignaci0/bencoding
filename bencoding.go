/*
This package shall be used to encode and decode
arbitrary go types into Bencoding format
*/
package bencoding

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
)

/*
Buffer shall be used to write into or read from for
coding/decoding data
*/
type Buffer struct {
	b bytes.Buffer
}

/*
This function shall retrieve the string representation of the current buffer
The buffer shall be never incomplete
*/
func (b *Buffer) String() string {
	return b.b.String()
}

func (b *Buffer) Bytes() []byte {
	return b.b.Bytes()
}

/*
Appends an existing bencoded string into the given buffer
*/
func (b *Buffer) WriteEncoded(bencoded string) {
	b.b.Write([]byte(bencoded))
}

/*
Write function encodes the given values into a
bencoding.Buffer.
Supported types are:

* string: for UTF-8 strings

* int: for numbers (TODO: give int64 support)

* []interface{}: slices of interfaces for lists

* map[string]interface{}: a map of interfaces indexed by strings

If the type is compounded (list or map) it shall encode its elements
as well.

*/
func (b *Buffer) Write(d interface{}) error {
	switch v := reflect.ValueOf(d); v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.b.WriteString(fmt.Sprintf("i%de", d))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.b.WriteString(fmt.Sprintf("i%de", d))

	case reflect.String:
		b.b.WriteString(fmt.Sprintf("%d:%s", len(v.String()), v.String()))

	case reflect.Slice, reflect.Array:

		b.b.WriteString("l")
		for i := 0; i < v.Len(); i++ {
			b.Write(v.Index(i).Interface())
		}
		b.b.WriteString("e")

	case reflect.Map:

		var sortedKeys []string
		for _, x := range v.MapKeys() {
			//This works as a sanity check; as per specification,
			//map keys must be strings
			if x.Kind() != reflect.String {
				return errors.New("Invalid map")
			}
			sortedKeys = append(sortedKeys, x.String())
		}

		sort.Strings(sortedKeys)

		b.b.WriteString("d")
		for _, k := range sortedKeys { //There has to be a better way to do this
			b.Write(k) //This key was converted (forced to string)
			b.Write(v.MapIndex(reflect.ValueOf(k)).Interface())
		}
		b.b.WriteString("e")

	default:
		log.Println("bencoding: Unhandled type: ", d)
		b.b.WriteString("le")
	}

	return nil
}

/*
Read shall retrieve the next complete value from the current
buffer and it shall retrieve nil on buffer depletion

It might return an error when coded data is invalid or incomplete
*/
func (b *Buffer) Read() (interface{}, error) {
	var i string
Loop:
	for {
		char, err := b.b.ReadByte()

		if err == io.EOF {
			return nil, nil
		}

		switch {
		case char == 'i':
			i = "i"
			break Loop
		case char == 'd':
			i = "d"
			break Loop
		case char == 'l':
			i = "l"
			break Loop
		case char >= '0' && char <= '9':
			i = fmt.Sprintf("%s%c", i, char)
		case char == ':':
			break Loop
		default:
			//I guess this is an e, let's send it back
			if char != 'e' {
				return nil, errors.New("Invalid encoding")
			}
			b.b.UnreadByte()
			break Loop
		}
	}

	if i == "i" {
		var rv int64
		str, err := b.b.ReadString('e')

		if err != nil {
			return nil, errors.New("Invalid coding")
		}

		fmt.Sscanf(str, "%de", &rv)

		return rv, nil
	} else if i == "d" {
		m := make(map[string]interface{})

		for char, err := b.b.ReadByte(); char != 'e'; {
			if err != nil {
				return nil, errors.New("Truncated coding")
			}

			b.b.UnreadByte()

			key, _ := b.Read()
			value, _ := b.Read()
			m[key.(string)] = value
			char, err = b.b.ReadByte()
		}

		return m, nil

	} else if i == "l" {
		var list []interface{}
		for {
			c, err := b.b.ReadByte()
			if c == 'e' {
				break
			} else if err == io.EOF {
				return nil, errors.New("Invalid list received")
			}
			b.b.UnreadByte()

			element, err := b.Read()

			if err != nil {
				return nil, err
			}

			list = append(list, element)
		}

		return list, nil
	} else { //it's a string
		var len int
		fmt.Sscanf(i, "%d", &len)

		var str = make([]byte, len)
		n, _ := b.b.Read(str)
		if n < len {
			return nil, errors.New("String is not long enough")
		}

		return string(str), nil
	}
}
