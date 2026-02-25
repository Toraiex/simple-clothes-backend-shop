package handler

import (
	"simple-clothes-shop/internal/domain"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type UserHandler struct {
	service domain.UserService
}

type VerifyEmailInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (h *UserHandler) Register(c *fiber.Ctx) error {

	type RegisterInput struct {
		Username string `json:"username" validate:"required,min=6"` // บังคับกรอก, ขั้นต่ำ 6 ตัว
		Password string `json:"password" validate:"required,min=6"` // บังคับกรอก, ขั้นต่ำ 6 ตัว
		Address  string `json:"address"`                            // ไม่บังคับ
		Phone    string `json:"phone" validate:"required,len=10"`   // บังคับกรอก, ต้องยาว 10 ตัวเป๊ะ
		Email    string `json:"email" validate:"required,email"`    // บังคับกรอก, ต้องเป็นฟอร์แมตอีเมล
	}

	var input RegisterInput

	// ดึงข้อมูลจาก Body มาใส่ Struct
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// ✅ 4. สั่งตรวจสอบข้อมูลตามกฎที่เราเขียนไว้ใน Tag
	if err := validate.Struct(&input); err != nil {
		// ถ้าข้อมูลไม่ตรงเงื่อนไข ให้เตะกลับไปพร้อมบอกว่าฟิลด์ไหนผิด
		return c.Status(400).JSON(fiber.Map{
			"error":   "ข้อมูลไม่ผ่านเกณฑ์การตรวจสอบ (เช่น อีเมลผิดรูปแบบ หรือ รหัสผ่านสั้นเกินไป)",
			"details": err.Error(),
		})
	}

	// ✅ 5. ถ้าข้อมูลเป๊ะหมด ค่อยประกอบร่างส่งให้ Service (โค้ดส่วนนี้ของคุณเขียนดีแล้วครับ)
	user := domain.User{
		Username: input.Username,
		Password: input.Password,
		Address:  input.Address,
		Phone:    input.Phone,
		Email:    input.Email,
	}

	if err := h.service.Register(&user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ",
		"user":    user,
	})
}

// ==========================================
// 2. เข้าสู่ระบบ (Login) - เวอร์ชัน Pro
// ==========================================
func (h *UserHandler) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// ✅ 1. ให้ Fiber ดึงข้อมูลอุปกรณ์ (User-Agent) และ IP Address ของลูกค้าให้
	userAgent := c.Get("User-Agent")
	clientIP := c.IP()

	// ✅ 2. ส่งข้อมูลทั้งหมดให้ Service ทำงานต่อ
	accessToken, refreshToken, role, err := h.service.Login(input.Username, input.Password, userAgent, clientIP)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// สร้าง HttpOnly Cookie (เหมือนที่ทำในสเต็ปที่แล้วเป๊ะเลยครับ)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	})

	return c.JSON(fiber.Map{
		"message": "เข้าสู่ระบบสำเร็จ",
		"role":    role,
	})
}

// ==========================================
// 3. ต่ออายุ Token (Refresh)
// ==========================================
func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: ไม่พบ Refresh Token"})
	}

	// ✅ รับ Token ใบใหม่มา 2 ใบ
	newAccessToken, newRefreshToken, err := h.service.RefreshAccessToken(refreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	// ✅ อบคุกกี้ Access Token ใหม่ (ทับของเดิม)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	})

	// ✅ อบคุกกี้ Refresh Token ใหม่ (ทับของเดิม)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour), // ยืดอายุคุกกี้ไปอีก 7 วัน
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	})

	return c.JSON(fiber.Map{
		"message": "ต่ออายุ Token และเซสชันสำเร็จ",
	})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	idParam, _ := strconv.Atoi(c.Params("id"))

	// ✅ ดึงค่าออกมาเป็น uint ตรงๆ เพราะตอนเซฟใน Middleware เราเซฟเป็น uint ไปแล้ว
	requesterID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// ✅ ดึงค่า Role
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	requesterRole := domain.Role(roleStr)

	user, err := h.service.GetUser(requesterID, requesterRole, uint(idParam))
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	user.Password = ""
	return c.JSON(user)
}

// ==========================================
// ดูรายชื่อผู้ใช้งานทั้งหมด (GetAllUsers)
// ==========================================
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	users, err := h.service.GetAllUsers(domain.Role(roleStr))
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลผู้ใช้งานสำเร็จ",
		"data":    users,
	})
}

// ==========================================
// อัปเดตข้อมูลผู้ใช้ (UpdateUser)
// ==========================================
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	idParam, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ผู้ใช้งานไม่ถูกต้อง"})
	}

	requesterID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	requesterRole := domain.Role(roleStr)

	// สร้าง DTO เพื่อบังคับให้ Client ส่งมาได้แค่ 3 ฟิลด์นี้เท่านั้น (ป้องกันคนเนียนส่ง Password มาแก้)
	type UpdateUserInput struct {
		Address string `json:"address"`
		Phone   string `json:"phone"`
		Role    string `json:"role"`
	}

	var input UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูล JSON ไม่ถูกต้อง"})
	}

	// นำเข้า Service
	err = h.service.UpdateUser(
		requesterID,
		requesterRole,
		uint(idParam),
		&domain.User{
			Address: input.Address,
			Phone:   input.Phone,
			Role:    domain.Role(input.Role), // แปลงเป็น type Role ก่อนส่ง
		},
	)

	if err != nil {
		// จับ Error ถ้าเป็น 404 ไม่พบผู้ใช้ หรือ 403 Forbidden
		if err.Error() == "ไม่พบข้อมูลผู้ใช้งานนี้ในระบบ" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตข้อมูลผู้ใช้งานสำเร็จ"})
}

// ==========================================
// 4. ออกจากระบบ (Logout)
// ==========================================
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	// 1. ดึง Refresh Token มาเช็ค
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		// 2. ส่งให้ Service ไปจัดการบล็อกใน Database
		_ = h.service.Logout(refreshToken)
	}

	// 3. ท่าไม้ตายทำลายคุกกี้: สั่งเซ็ตค่าให้ว่างเปล่า และตั้งเวลาหมดอายุเป็น "อดีต" (-1 ชั่วโมง)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "ออกจากระบบสำเร็จ",
	})
}
func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	var input VerifyEmailInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	err := h.service.VerifyEmail(input.Email, input.OTP)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "ยินดีด้วย! ยืนยันอีเมลสำเร็จแล้ว ตอนนี้คุณสามารถเข้าสู่ระบบได้เต็มรูปแบบ",
	})
}

// ใน user_handler.go
func (h *UserHandler) ResendOTP(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	err := h.service.ResendOTP(input.Email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP ใหม่ไปที่อีเมลของคุณแล้ว"})
}
