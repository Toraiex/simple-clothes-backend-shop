package http

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/middleware"
	"simple-clothes-shop/pkg/utils"

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

	userGroup := router.Group("/users")
	authGroup := router.Group("/auth")

	// --- Public Routes ---
	authGroup.Post("/register", handler.Register)
	authGroup.Post("/login", authLimiter, handler.Login)
	authGroup.Post("/refresh", handler.RefreshToken)
	authGroup.Post("/verify-email", handler.VerifyEmail)
	authGroup.Post("/resend-otp", authLimiter, handler.ResendOTP)
	authGroup.Post("/forgot-password", handler.ForgotPassword)
	authGroup.Post("/resend-reset-otp", authLimiter, handler.ResendResetOTP)
	authGroup.Post("/reset-password", handler.ResetPassword)

	// --- Protected Routes ---
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
		err = fmt.Errorf("ข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := validate.Struct(&input); err != nil {
		err = fmt.Errorf("ข้อมูลไม่ผ่านเกณฑ์: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if !isComplexPassword(input.Password) {
		err := fmt.Errorf("รหัสผ่านไม่ปลอดภัยพอ ต้องมีตัวพิมพ์ใหญ่/เล็ก/ตัวเลข/สัญลักษณ์: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	user := domain.User{
		Username: input.Username,
		Password: input.Password,
		Address:  input.Address,
		Phone:    input.Phone,
		Email:    input.Email,
	}

	if err := h.usecase.Register(c.UserContext(), &user); err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "สมัครสมาชิกสำเร็จ"})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("ข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	accessToken, refreshToken, role, err := h.usecase.Login(c.UserContext(), input.Username, input.Password)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	h.setAuthCookies(c, accessToken, refreshToken)
	return c.JSON(fiber.Map{"message": "เข้าสู่ระบบสำเร็จ", "role": role})
}

func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		err := fmt.Errorf("ไม่พบ Refresh Token: %w", domain.ErrUnauthorized)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	newAccessToken, newRefreshToken, err := h.usecase.RefreshAccessToken(c.UserContext(), refreshToken)
	if err != nil {
		h.clearAuthCookies(c)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	h.setAuthCookies(c, newAccessToken, newRefreshToken)
	return c.JSON(fiber.Map{"message": "ต่ออายุ Token และเซสชันสำเร็จ"})
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// 1. ดึงข้อมูลคนรีเควสต์จาก Context ก่อน
	requesterID, err := middleware.GetUserID(c)
	if err != nil {
		err := fmt.Errorf("เซสชันไม่ถูกต้อง: %w", domain.ErrUnauthorized)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	roleStr, err := middleware.GetUserRole(c)
	if err != nil {
		err := fmt.Errorf("เซสชันไม่ถูกต้อง: %w", domain.ErrUnauthorized)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}
	requesterRole := domain.Role(roleStr)

	paramID := c.Params("id")
	var targetID uint

	if paramID == "me" {
		targetID = requesterID
	} else {

		parsedID, err := strconv.Atoi(paramID)
		if err != nil {
			err = fmt.Errorf("รูปแบบ ID ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
			return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
		}
		targetID = uint(parsedID)
	}

	user, err := h.usecase.GetUser(c.UserContext(), requesterID, requesterRole, targetID)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	user.Password = ""
	return c.JSON(user)
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		err := fmt.Errorf("เซสชันไม่ถูกต้อง: %w", domain.ErrUnauthorized)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	users, err := h.usecase.GetAllUsers(c.UserContext(), domain.Role(roleStr))
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ดึงข้อมูลผู้ใช้งานสำเร็จ", "data": users})
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// 💡 1. ดึง ID และ Role ด้วย Helper ของเรา (ปลอดภัย 100%)
	requesterID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	roleStr, err := middleware.GetUserRole(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}
	requesterRole := domain.Role(roleStr)

	// 💡 2. จัดการกับ Params ID (รองรับท่า /users/me ด้วย)
	paramID := c.Params("id")
	var targetID uint

	if paramID == "me" {
		targetID = requesterID
	} else {
		parsedID, err := strconv.Atoi(paramID)
		if err != nil {
			err = fmt.Errorf("รูปแบบ ID ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
			return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
		}
		targetID = uint(parsedID)
	}

	// 3. รับและตรวจสอบข้อมูลจาก Body
	type UpdateInput struct {
		Address *string `json:"address"`
		Phone   *string `json:"phone" validate:"omitempty,len=10,numeric"`
		Role    *string `json:"role"`
	}

	var input UpdateInput
	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบ JSON ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	// (สมมติว่าคุณมีตัวแปร validate ประกาศไว้แล้วระดับ Global ของ package นี้)
	if err := validate.Struct(&input); err != nil {
		err = fmt.Errorf("ข้อมูลไม่ตรงตามเงื่อนไข: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	// 4. เตรียมข้อมูลสำหรับอัปเดต
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

	// 5. ส่งให้ Usecase จัดการ
	err = h.usecase.UpdateUser(c.UserContext(), requesterID, requesterRole, targetID, updateData)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตข้อมูลสำเร็จ"})
}

func (h *UserHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.usecase.Logout(c.UserContext(), refreshToken)
	}

	h.clearAuthCookies(c)
	return c.JSON(fiber.Map{"message": "ออกจากระบบสำเร็จ"})
}

func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	var input VerifyEmailInput
	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.VerifyEmail(c.UserContext(), input.Email, input.OTP); err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ยินดีด้วย! ยืนยันอีเมลสำเร็จแล้ว ตอนนี้คุณสามารถเข้าสู่ระบบได้เต็มรูปแบบ"})
}

func (h *UserHandler) ResendOTP(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.ResendOTP(c.UserContext(), input.Email); err != nil {

		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP ใหม่ไปที่อีเมลของคุณแล้ว"})
}

func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.ForgotPassword(c.UserContext(), input.Email); err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ส่งรหัส OTP สำหรับตั้งรหัสผ่านใหม่ ไปที่อีเมลของคุณแล้ว"})
}

func (h *UserHandler) ResendResetOTP(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}
	if err := validate.Struct(&input); err != nil {
		err = fmt.Errorf("ข้อมูลไม่ถูกต้องตามรูปแบบ: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.ResendResetOTP(c.UserContext(), input.Email); err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
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
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := validate.Struct(&input); err != nil {
		err = fmt.Errorf("ข้อมูลไม่ถูกต้องตามรูปแบบ: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if !isComplexPassword(input.NewPassword) {
		err := fmt.Errorf("รหัสผ่านใหม่ต้องมีตัวพิมพ์ใหญ่, ตัวเล็ก, ตัวเลข และสัญลักษณ์: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.ResetPassword(c.UserContext(), input.Email, input.OTP, input.NewPassword); err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "รีเซ็ตรหัสผ่านสำเร็จ! คุณสามารถเข้าสู่ระบบด้วยรหัสผ่านใหม่ได้ทันที"})
}

func isComplexPassword(pass string) bool {
	return regexLower.MatchString(pass) &&
		regexUpper.MatchString(pass) &&
		regexNumber.MatchString(pass) &&
		regexSpecial.MatchString(pass)
}

func (h *UserHandler) setAuthCookies(c *fiber.Ctx, accessToken string, refreshToken string) {
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
}

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

// 💡 สไตล์ bxcodec: แปลง Error มาตรฐานเป็น Status Code
