package service

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// สร้างแบบแปลนสำหรับบุรุษไปรษณีย์
type EmailService interface {
	SendVerificationEmail(toEmail string, otpCode string) error
}

type emailService struct{}

// ฟังก์ชันจ้างบุรุษไปรษณีย์คนใหม่
func NewEmailService() EmailService {
	return &emailService{}
}

// ==========================================
// ฟังก์ชันส่งอีเมลพร้อมรหัส OTP
// ==========================================
func (s *emailService) SendVerificationEmail(toEmail string, otpCode string) error {
	m := gomail.NewMessage()

	// 1. ตั้งค่าหัวจดหมาย (ใครส่งหาใคร)
	m.SetHeader("From", os.Getenv("SMTP_EMAIL"))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "รหัสยืนยันอีเมลของคุณจาก Simple Clothes Shop 👕")

	// 2. เนื้อหาในจดหมาย (ใช้ HTML ตกแต่งให้สวยงามได้)
	htmlBody := `
		<h2>ยินดีต้อนรับสู่ร้าน Simple Clothes Shop ครับ!</h2>
		<p>กรุณานำรหัส OTP ด้านล่างนี้ไปกรอกเพื่อยืนยันบัญชีของคุณ:</p>
		<h1 style="color: #4CAF50; background-color: #f4f4f4; padding: 10px; text-align: center; border-radius: 5px;">` + otpCode + `</h1>
		<p style="color: red;">*รหัสนี้มีอายุการใช้งาน 15 นาที</p>
	`
	m.SetBody("text/html", htmlBody)

	// 3. ตั้งค่าการเชื่อมต่อ (ดึงข้อมูลจาก .env)
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT")) // แปลง Port จาก String เป็นตัวเลข
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_EMAIL"),
		os.Getenv("SMTP_PASSWORD"),
	)

	// 4. สั่งส่งออกไปเลย!
	return d.DialAndSend(m)
}
