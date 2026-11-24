// Package pyjson reads JSON keeping key order and number literals, and writes it back the way Python's json.dumps does.
package pyjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var errTrailingData = errors.New("data after the JSON value")

// Value is nil, bool, string, Number, int64, []Value or *Object.
type Value = any

// Number keeps the literal, which Python reads as an int or a float and prints accordingly.
type Number string

// Object keeps its keys in the order they were first set, like a Python dict.
type Object struct {
	keys   []string
	values map[string]Value
}

func NewObject() *Object {
	return &Object{values: map[string]Value{}}
}

func (object *Object) Set(key string, value Value) *Object {
	if _, exists := object.values[key]; !exists {
		object.keys = append(object.keys, key)
	}
	object.values[key] = value
	return object
}

func (object *Object) Get(key string) (Value, bool) {
	value, ok := object.values[key]
	return value, ok
}

// Decode reads one JSON value; a duplicate key keeps its first position and its last value, as json.loads does.
func Decode(data []byte) (Value, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errTrailingData
	}
	return value, nil
}

func decodeValue(decoder *json.Decoder) (Value, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("read JSON: %w", err)
	}
	switch token := token.(type) {
	case json.Delim:
		if token == '[' {
			return decodeArray(decoder)
		}
		return decodeObject(decoder)
	case json.Number:
		return Number(token), nil
	default:
		return token, nil
	}
}

func decodeArray(decoder *json.Decoder) (Value, error) {
	items := []Value{}
	for decoder.More() {
		item, err := decodeValue(decoder)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	_, err := decoder.Token()
	return items, err
}

func decodeObject(decoder *json.Decoder) (Value, error) {
	object := NewObject()
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("read JSON object key: %w", err)
		}
		value, err := decodeValue(decoder)
		if err != nil {
			return nil, err
		}
		object.Set(key.(string), value) //nolint:forcetypeassert // reason: encoding/json only yields string keys inside an object
	}
	_, err := decoder.Token()
	return object, err
}
