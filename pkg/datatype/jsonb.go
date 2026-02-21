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

// Scan: แปลงข้อมูลที่อ่านจาก Database (JSONB) ให้กลายเป็น Go Array
func (a *JSONStringArray) Scan(value interface{}) error {
	// ดักไว้ก่อน: ถ้าใน DB เป็นค่าว่าง (null) ให้เซ็ตเป็น Array เปล่าๆ ป้องกันแอปพัง
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

// Value: แปลงจาก Go Array ให้กลายเป็น JSONB เพื่อเซฟลง Database
func (a JSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]string{}) // ป้องกันการบันทึกค่า null ลง DB
	}
	return json.Marshal(a)
}

// ----------------------------------------------------
// 2. JSONAttributes (สำหรับเก็บตัวเลือกสินค้าแบบยืดหยุ่น)
// ----------------------------------------------------
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
