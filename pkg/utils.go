package handler

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ✅ 1. ย้าย Regex ออกมาคอมไพล์ไว้ล่วงหน้าแค่ครั้งเดียวตอนเริ่มโปรแกรม
var (
	regexLower   = regexp.MustCompile(`[a-z]`)
	regexUpper   = regexp.MustCompile(`[A-Z]`)
	regexNumber  = regexp.MustCompile(`[0-9]`)
	regexSpecial = regexp.MustCompile(`[!@#$%^&*]`)
)

func isComplexPassword(pass string) bool {
	// กฎ: ตัวเล็กอย่างน้อยหนึ่ง, ตัวใหญ่อย่างน้อยหนึ่ง, ตัวเลขอย่างน้อยหนึ่ง, สัญลักษณ์อย่างน้อยหนึ่ง
	var (
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(pass)
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(pass)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(pass)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*]`).MatchString(pass)
	)
	return hasLower && hasUpper && hasNumber && hasSpecial
}
