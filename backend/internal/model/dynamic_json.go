package model

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
)

// DynamicJSON is a reusable type that holds raw JSON bytes
type DynamicJSON json.RawMessage

// Scan reads from PostgreSQL (JSONB) into your Go struct
func (j *DynamicJSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// CRITICAL: Make a copy of the byte slice!
		// Database drivers often reuse the memory buffer for []byte.
		b := make([]byte, len(v))
		copy(b, v)
		*j = b
		return nil
	case string:
		*j = []byte(v)
		return nil
	default:
		return errors.New("type assertion failed for DynamicJSON")
	}
}

// Value writes from your Go struct into PostgreSQL (JSONB)
func (j DynamicJSON) Value() (driver.Value, error) {
	// If the JSON is empty, default to an empty JSON object "{}" to avoid DB errors
	if len(j) == 0 {
		return "{}", nil
	}

	// Convert the raw bytes to a string for the Postgres driver to insert
	return string(j), nil
}

// MarshalJSON ensures the bytes are output as raw JSON objects, not Base64 strings
func (j DynamicJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	// Cast back to json.RawMessage to use the standard library's JSON logic
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON ensures that when your frontend sends JSON back, Go reads it correctly
func (j *DynamicJSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("DynamicJSON: UnmarshalJSON on nil pointer")
	}
	return (*json.RawMessage)(j).UnmarshalJSON(data)
}

func StructToDynamicJSON(data any) (DynamicJSON, error) {
	if data == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Cast the bytes to your DynamicJSON type
	return DynamicJSON(bytes), nil
}

func (j *DynamicJSON) AreJSONsEqual(b DynamicJSON) bool {
	if bytes.Equal(*j, b) {
		return true
	}

	// 2. Slow path: parse & compare semantically
	var objA, objB any
	if json.Unmarshal(*j, &objA) != nil || json.Unmarshal(b, &objB) != nil {
		return false
	}
	return reflect.DeepEqual(objA, objB)
}
