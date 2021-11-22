package bencoding

import (
	"testing"
)

func TestWriteEncoded(test *testing.T) {
	var b Buffer
	b.Write([]byte("i3e"))

	str := b.String()
	if str != "i3e" {
		test.Errorf("Expected 'i3e', got '%s'", str)
	}
}

func TestReadInteger(test *testing.T) {

	var b Buffer
	b.Write([]byte("i3e"))

	got, _ := b.Decode()
	if got != int64(3) {
		test.Errorf("Expected 3, got %d", got)
	}

	got, err := b.Decode()
	if got != nil || err != nil {
		test.Errorf("Expected nil, got %v", got)
	}
}

func TestReadString(test *testing.T) {

	var b Buffer
	b.Write([]byte("3:cow"))

	got, _ := b.Decode()
	if got != "cow" {
		test.Errorf("Expected 'cow', got %s", got)
	}
}

func TestReadList(test *testing.T) {
	var b Buffer
	expected := []string{"cow", "01234567891"}
	b.Write([]byte("l3:cow11:01234567891e"))

	got, err := b.Decode()
	if err != nil {
		test.Errorf(err.Error())
	}

	switch val := got.(type) {
	default:
		test.Errorf("Expected a slice, got something else")
		break
	case []interface{}:
		for k, v := range expected {
			if val[k] != v {
				test.Errorf("Expected %s, got %s", v, val[k])
			}
		}
		break
	}
}

func TestWriteList(test *testing.T) {
	var b Buffer
	var expected string = "li3e3:cowe"
	var list []interface{} = []interface{}{3, "cow"}

	b.Encode(list)
	var got string = b.String()
	if got != expected {
		test.Errorf("Expected %s, got %s", expected, got)
	}

}

func TestInvalidMap(test *testing.T) {
	var b Buffer

	err := b.Encode(map[int]string{1: "some value"})
	if err != ErrorInvalidMap {
		test.Errorf("Expected %v, got %v", ErrorInvalidMap, err)
	}
}

func TestInvalidType(test *testing.T) {
	var b Buffer
	type st struct {
		attribute string
	}
	err := b.Encode(st{"hello"})
	if err != ErrorInvalidType {
		test.Errorf("Expected %v, got %v", ErrorInvalidType, err)
	}
}

func TestReadMap(test *testing.T) {
	var b Buffer
	b.Write([]byte("d3:cow3:doge"))
	got, err := b.Decode()
	if err != nil {
		test.Errorf("Got error %+v", err)
	}

	m := got.(map[string]interface{})
	if m["cow"].(string) != "dog" {
		test.Errorf("Expected map with cow:dog and got %+v", m)
	}
}

func TestWriteInteger(test *testing.T) {
	var b Buffer

	b.Encode(3)
	got := b.String()
	if got != "i3e" {
		test.Errorf("Expected i3e; Got %+v", got)
	}
}

func TestWriteString(test *testing.T) {
	var b Buffer

	b.Encode("hello")
	got := b.String()
	if got != "5:hello" {
		test.Errorf("Expected 5:hello; Got %+v", got)
	}
}

func TestWriteUInt16(test *testing.T) {
	var b Buffer

	b.Encode(uint16(1))
	got := b.String()
	expected := "i1e"
	if got != expected {
		test.Errorf("Expected %s; Got %+v", expected, got)
	}
}
func TestMapComplex(test *testing.T) {
	r := make(map[string]map[string]interface{})
	r["files"] = make(map[string]interface{})
	r["files"]["info_hash_2"] = map[string]interface{}{
		"complete":   0,
		"downloaded": 1,
		"incomplete": 2,
	}
	r["files"]["info_hash_1"] = map[string]interface{}{
		"complete":   3,
		"downloaded": 4,
		"incomplete": 3,
	}
	var buffer Buffer
	buffer.Encode(r)
	expected := "d5:filesd11:info_hash_1d8:completei3e10:downloadedi4e10:incompletei3ee11:info_hash_2d8:completei0e10:downloadedi1e10:incompletei2eeee"
	got := buffer.String()

	if expected != got {
		test.Errorf("Expected %v. Got: %v", expected, got)
	}
}
