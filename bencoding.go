/*
This package shall be used to encode and decode
arbitrary go types into Bencoding format

The bencoding.Buffer implements:
	io.Reader
	fmt.Stringer
*/
package bencoding

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
)

/*
Buffer shall be used to write into or read from for
coding/decoding data

It extends the bytes.Buffer type allowing its operations
and the provided by the interfaces bencoding.Decoder and
bencoding.Encoder
*/
type Buffer struct {
	bytes.Buffer
}

type Decoder interface {
	Decode() (interface{}, error)
}

type Encoder interface {
	Encode(interface{}) error
}

var (
	ErrorInvalidMap      error = errors.New("Invalid map")
	ErrorInvalidType     error = errors.New("Unhandled type")
	ErrorInvalidData     error = errors.New("Invalid data in buffer")
	ErrorInvalidList     error = errors.New("Invalid list received")
	ErrorTruncatedString error = errors.New("String is not long enough")
)

/*
Write function encodes the given values into a bencoding.Buffer.

Supported value's types are:

	string: for UTF-8 strings
	int: for numbers (TODO: give int64 support)
	[]interface{}: slices of interfaces for lists
	map[string]interface{}: a map of interfaces indexed by strings

If the type is compounded (list or map) it shall encode its elements
as well.
*/
func (b *Buffer) Encode(d interface{}) error {
	switch v := reflect.ValueOf(d); v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b.WriteString(fmt.Sprintf("i%de", d))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(fmt.Sprintf("i%de", d))

	case reflect.String:
		b.WriteString(fmt.Sprintf("%d:%s", len(v.String()), v.String()))

	case reflect.Slice, reflect.Array:

		b.WriteString("l")
		for i := 0; i < v.Len(); i++ {
			b.Encode(v.Index(i).Interface())
		}
		b.WriteString("e")

	case reflect.Map:

		var sortedKeys []string
		for _, x := range v.MapKeys() {
			//This works as a sanity check; as per specification,
			//map keys must be strings
			if x.Kind() != reflect.String {
				return ErrorInvalidMap
			}
			sortedKeys = append(sortedKeys, x.String())
		}

		sort.Strings(sortedKeys)

		b.WriteString("d")
		for _, k := range sortedKeys { //There has to be a better way to do this
			b.Encode(k) //This key was converted (forced to string)
			b.Encode(v.MapIndex(reflect.ValueOf(k)).Interface())
		}
		b.WriteString("e")

	default:
		return ErrorInvalidType
	}

	return nil
}

/*
Read shall retrieve the next complete value from the current
buffer and it shall retrieve nil on buffer depletion

It might return an error when coded data is invalid or incomplete
*/
func (b *Buffer) Decode() (interface{}, error) {
	var i string
Loop:
	for {
		char, err := b.ReadByte()

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
				return nil, ErrorInvalidData
			}
			b.UnreadByte()
			break Loop
		}
	}

	if i == "i" {
		var rv int64
		str, err := b.ReadString('e')

		if err != nil {
			return nil, ErrorInvalidData
		}

		fmt.Sscanf(str, "%de", &rv)

		return rv, nil
	} else if i == "d" {
		m := make(map[string]interface{})

		for char, err := b.ReadByte(); char != 'e'; {
			if err != nil {
				return nil, ErrorInvalidData
			}

			b.UnreadByte()

			key, _ := b.Decode()
			value, _ := b.Decode()
			m[key.(string)] = value
			char, err = b.ReadByte()
		}

		return m, nil

	} else if i == "l" {
		var list []interface{}
		for {
			c, err := b.ReadByte()
			if c == 'e' {
				break
			} else if err == io.EOF {
				return nil, ErrorInvalidList
			}
			b.UnreadByte()

			element, err := b.Decode()

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
		n, _ := b.Read(str) //When I get go 1.13 I might wrap the error
		if n < len {
			return nil, ErrorTruncatedString
		}

		return string(str), nil
	}
}
