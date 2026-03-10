package datatype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ----------------------------------------------------
// 1. JSONStringArray (สำหรับเก็บ Array ของรูปภาพ)
// ----------------------------------------------------
type JSONStringArray []string

func (a *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, a)
}

func (a JSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(a)
}

type JSONAttributes map[string]string

// Scan: DB -> Go
func (m *JSONAttributes) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]string)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}

// Value: Go -> DB
func (m JSONAttributes) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(make(map[string]string))
	}
	return json.Marshal(m)
}
