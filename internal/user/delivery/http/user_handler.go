package http

import (
	"regexp"
	"simple-clothes-shop/internal/domain"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

var (
	regexLower   = regexp.MustCompile(`[a-z]`)
	regexUpper   = regexp.MustCompile(`[A-Z]`)
	regexNumber  = regexp.MustCompile(`[0-9]`)
	regexSpecial = regexp.MustCompile(`[!@#$%^&*]`)
)

type UserHandler struct {
	usecase domain.UserUsecase
}

func NewUserHandler(router fiber.Router, usecase domain.UserUsecase, authMid fiber.Handler, adminMid fiber.Handler, authLimiter fiber.Handler) {
	handler := &UserHandler{usecase: usecase}

	// 📦 จัดกลุ่ม Route ที่เกี่ยวกับ User ทั้งหมด
	userGroup := router.Group("/users")
	authGroup := router.Group("/auth") // ถ้าอยากแยก auth ออกมาให้ดูคลีนขึ้น

	// --- Public Routes (ไม่ต้องล็อกอิน) ---
	authGroup.Post("/register", handler.Register)
	authGroup.Post("/login", authLimiter, handler.Login)
	authGroup.Post("/refresh", handler.RefreshToken)
	authGroup.Post("/verify-email", handler.VerifyEmail)
	authGroup.Post("/resend-otp", authLimiter, handler.ResendOTP)
	authGroup.Post("/forgot-password", handler.ForgotPassword)
	authGroup.Post("/resend-reset-otp", authLimiter, handler.ResendResetOTP)
	authGroup.Post("/reset-password", handler.ResetPassword)

	// --- Protected Routes (ต้องล็อกอิน) ---
	authGroup.Post("/logout", authMid, handler.Logout)

	userGroup.Get("/:id", authMid, handler.GetUser)
	userGroup.Patch("/:id", authMid, handler.UpdateUser)

	// --- Admin Routes ---
	userGroup.Get("/", authMid, adminMid, handler.GetAllUsers)
}

type VerifyEmailInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

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

	if err := h.usecase.Register(c.UserContext(), &user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ",
		"user":    user,
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	accessToken, refreshToken, role, err := h.usecase.Login(c.UserContext(), input.Username, input.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

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

func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: ไม่พบ Refresh Token"})
	}

	// ✅ รับ Token ใบใหม่มา 2 ใบ
	newAccessToken, newRefreshToken, err := h.usecase.RefreshAccessToken(c.UserContext(), refreshToken)
	if err != nil {
		h.clearAuthCookies(c) // ท่าไม้ตายเคลียร์คุกกี้ถ้า Refresh Token พัง
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	})

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

	// 🚀 เพิ่ม c.UserContext() เป็นพารามิเตอร์แรกสุด
	user, err := h.usecase.GetUser(c.UserContext(), requesterID, requesterRole, uint(idParam))
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	user.Password = ""
	return c.JSON(user)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// 🚀 เพิ่ม c.UserContext() เป็นพารามิเตอร์แรกสุด
	users, err := h.usecase.GetAllUsers(c.UserContext(), domain.Role(roleStr))
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

	// 🚀 เพิ่ม c.UserContext() เป็นพารามิเตอร์แรกสุด
	err := h.usecase.UpdateUser(c.UserContext(), requesterID, requesterRole, uint(idParam), updateData)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตข้อมูลสำเร็จ"})
}
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		// 🚀 แก้ไข: เพิ่ม c.UserContext()
		_ = h.usecase.Logout(c.UserContext(), refreshToken)
	}

	h.clearAuthCookies(c) // เรียกใช้ Helper ให้โค้ดสั้นลง

	return c.JSON(fiber.Map{"message": "ออกจากระบบสำเร็จ"})
}

func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	var input VerifyEmailInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 🚀 แก้ไข: เพิ่ม c.UserContext()
	err := h.usecase.VerifyEmail(c.UserContext(), input.Email, input.OTP)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ยินดีด้วย! ยืนยันอีเมลสำเร็จแล้ว ตอนนี้คุณสามารถเข้าสู่ระบบได้เต็มรูปแบบ"})
}

// ใน user_handler.go
func (h *UserHandler) ResendOTP(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 🚀 แก้ไข: เพิ่ม c.UserContext()
	err := h.usecase.ResendOTP(c.UserContext(), input.Email)
	if err != nil {
		return c.Status(429).JSON(fiber.Map{"error": err.Error()}) // เปลี่ยนเป็น 429 Too Many Requests ให้สมจริง
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP ใหม่ไปที่อีเมลของคุณแล้ว"})
}

func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// 🚀 แก้ไข: เพิ่ม c.UserContext()
	err := h.usecase.ForgotPassword(c.UserContext(), input.Email)
	if err != nil {
		return c.Status(429).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP สำหรับตั้งรหัสผ่านใหม่ ไปที่อีเมลของคุณแล้ว"})
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

	// 🚀 แก้ไข: เพิ่ม c.UserContext()
	if err := h.usecase.ResendResetOTP(c.UserContext(), input.Email); err != nil {
		return c.Status(429).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP สำหรับรีเซ็ตรหัสผ่านใหม่ไปที่อีเมลของคุณแล้ว"})
}

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

	if !isComplexPassword(input.NewPassword) {
		return c.Status(400).JSON(fiber.Map{"error": "รหัสผ่านใหม่ต้องมีตัวพิมพ์ใหญ่, ตัวเล็ก, ตัวเลข และสัญลักษณ์"})
	}

	// 🚀 แก้ไข: เพิ่ม c.UserContext()
	if err := h.usecase.ResetPassword(c.UserContext(), input.Email, input.OTP, input.NewPassword); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "รีเซ็ตรหัสผ่านสำเร็จ! คุณสามารถเข้าสู่ระบบด้วยรหัสผ่านใหม่ได้ทันที"})
}
func isComplexPassword(pass string) bool {
	return regexLower.MatchString(pass) &&
		regexUpper.MatchString(pass) &&
		regexNumber.MatchString(pass) &&
		regexSpecial.MatchString(pass)
}

// ผมเพิ่มฟังก์ชันนี้ให้ครับ จะได้ไม่ต้องเขียนเคลียร์คุกกี้ซ้ำๆ
func (h *UserHandler) clearAuthCookies(c *fiber.Ctx) {
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
}
