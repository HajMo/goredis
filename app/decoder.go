package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Type byte

type Value struct {
	typ   Type
	bytes []byte
	array []Value
}

const (
	Integer = ':'
	String  = '+'
	Bulk    = '$'
	Array   = '*'
	Error   = '-'
)

type RESP struct {
	Type  Type
	Raw   []byte
	Data  []byte
	Count int
}

func Decode(stream *bufio.Reader) (Value, error) {
	dataType, err := stream.ReadByte()
	if err != nil {
		return Value{}, err
	}
	t := Type(dataType)

	switch t {
	case String:
		return decodeString(stream)
	case Bulk:
		return decodeBulk(stream)
	case Array:
		return decodeArray(stream)
	default:
		return Value{}, fmt.Errorf("Invalid RESP data type byte: %s", string(dataType))
	}
}

// converts Value to string
func (value Value) String() string {
	if value.typ == String || value.typ == Bulk {
		return string(value.bytes)
	}

	return ""
}

// converts Value to an array.
func (value Value) Array() []Value {
	if value.typ == Array {
		return value.array
	}

	return []Value{}
}

func decodeInteger(stream *bufio.Reader) (Value, error) {
	return Value{}, nil
}

func decodeString(stream *bufio.Reader) (Value, error) {
	bytes, err := readUntilCRLF(stream)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read string: %s", err)
	}

	return Value{
		typ:   String,
		bytes: bytes,
	}, nil
}

func decodeBulk(stream *bufio.Reader) (Value, error) {
	bytes, err := readUntilCRLF(stream)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string: %s", err)
	}

	count, err := strconv.Atoi(string(bytes))
	if err != nil {
		return Value{}, fmt.Errorf("failed to parse bulk string length: %s", err)
	}

	readedBytes := make([]byte, count+2)

	if _, err := io.ReadFull(stream, readedBytes); err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string contents: %s", err)
	}

	return Value{
		typ:   Bulk,
		bytes: readedBytes[:count],
	}, nil
}

func decodeArray(stream *bufio.Reader) (Value, error) {
	bytes, err := readUntilCRLF(stream)
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string length: %s", err)
	}

	count, err := strconv.Atoi(string(bytes))
	if err != nil {
		return Value{}, fmt.Errorf("failed to read bulk string length: %s", err)
	}

	array := []Value{}

	for i := 0; i < count; i++ {
		value, err := Decode(stream)
		if err != nil {
			return Value{}, err
		}

		array = append(array, value)
	}

	return Value{
		typ:   Array,
		array: array,
	}, nil
}

func readUntilCRLF(stream *bufio.Reader) ([]byte, error) {
	readedBytes := []byte{}

	for {
		b, err := stream.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		readedBytes = append(readedBytes, b...)

		if len(readedBytes) >= 2 && readedBytes[len(readedBytes)-2] == '\r' {
			break
		}
	}

	return readedBytes[:len(readedBytes)-2], nil
}
