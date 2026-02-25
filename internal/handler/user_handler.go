package handler

import (
	"regexp"
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
		Username string `json:"username" validate:"required,min=6,max=20"`
		Password string `json:"password" validate:"required,min=6,max=20"`
		Address  string `json:"address"`
		Phone    string `json:"phone" validate:"len=10,numeric"`
		Email    string `json:"email" validate:"required,email"`
	}

	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if err := validate.Struct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ผ่านเกณฑ์ (เช่น Username ต้องเป็นภาษาอังกฤษ/ตัวเลข 4 ตัวขึ้นไป)"})
	}

	if !isComplexPassword(input.Password) {
		return c.Status(400).JSON(fiber.Map{
			"error": "รหัสผ่านไม่ปลอดภัยพอ: ต้องมีตัวพิมพ์ใหญ่, ตัวพิมพ์เล็ก, ตัวเลข และสัญลักษณ์อย่างน้อย 1 ตัว",
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

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	idParam, _ := strconv.Atoi(c.Params("id"))
	requesterRole := domain.Role(c.Locals("role").(string))
	requesterID := c.Locals("user_id").(uint)

	type UpdateInput struct {
		Address *string `json:"address"`
		Phone   *string `json:"phone" validate:"omitempty,len=10,numeric"`
		Role    *string `json:"role"`
	}

	var input UpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "JSON ไม่ถูกต้อง"})
	}
	if err := validate.Struct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงค่าจาก pointer มาใส่ domain model
	updateData := &domain.User{}
	if input.Address != nil {
		updateData.Address = *input.Address
	}
	if input.Phone != nil {
		updateData.Phone = *input.Phone
	}
	if input.Role != nil {
		updateData.Role = domain.Role(*input.Role)
	}

	err := h.service.UpdateUser(requesterID, requesterRole, uint(idParam), updateData)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตข้อมูลสำเร็จ"})
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

// ==========================================
// 7. ขอรีเซ็ตรหัสผ่าน (Forgot Password)
// ==========================================
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"` // ถ้าใช้ validator ก็ใส่ tag ไว้ได้เลยครับ
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	err := h.service.ForgotPassword(input.Email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "ส่งรหัส OTP สำหรับตั้งรหัสผ่านใหม่ ไปที่อีเมลของคุณแล้ว",
	})
}

// 7.1 ขอส่งรหัส OTP สำหรับรีเซ็ตรหัสผ่านซ้ำ (Resend Reset OTP)
func (h *UserHandler) ResendResetOTP(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	if err := validate.Struct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้องตามรูปแบบ"})
	}

	if err := h.service.ResendResetOTP(input.Email); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "ส่งรหัส OTP สำหรับรีเซ็ตรหัสผ่านใหม่ไปที่อีเมลของคุณแล้ว",
	})
}

// ==========================================
// 8. ตั้งรหัสผ่านใหม่ (Reset Password)
// ==========================================
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Email       string `json:"email" validate:"required,email"`
		OTP         string `json:"otp" validate:"required,len=6"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	if err := validate.Struct(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้องตามรูปแบบ"})
	}

	// ✅ ตรวจสอบความปลอดภัยรหัสผ่านใหม่
	if !isComplexPassword(input.NewPassword) {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านใหม่ต้องมีตัวพิมพ์ใหญ่, ตัวเล็ก, ตัวเลข และสัญลักษณ์"})
	}

	if err := h.service.ResetPassword(input.Email, input.OTP, input.NewPassword); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "รีเซ็ตรหัสผ่านสำเร็จ! คุณสามารถเข้าสู่ระบบด้วยรหัสผ่านใหม่ได้ทันที",
	})
}

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
