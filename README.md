# bencoding

This module is just my _learning_ Go project.

It shouldn't be used or considered for use (or maybe it could). There are other good packages doing the same (and better) somewhere here.

The reason this is being published is I want to test the modules functionality

## Examples

To start encode, define your buffer variable as:

```go
  var buff bencoding.Buffer
```

With this you can operate the buffer with the existing interfaces plus the Encoder and Decoder interfaces

To encode an integer value:

```go
  buff.Encode(3)
  buff.Encode(int64(43243243223))
```

Strings must be passed as strings, note byte slices:

```go
  buff.Encode("Hello World!")
  buff.Encode(string([]byte{'h', 'e', 'l', 'l', 'o'}))
```

When passing arrays or slices, bencoded lists are generated:

```go
  buff.Encode([]interface{}{3, "hello", "World!", map[string]string{"key1": "value1", "key2": "value2"}})
```

If what's needed is to decode:

```go
  result, err := buff.Decode() //result is an interface{}
  if err != nil {
    //...
  }
```
