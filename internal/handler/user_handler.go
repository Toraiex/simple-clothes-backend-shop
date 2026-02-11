package handler

import (
	"simple-clothes-shop/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service domain.UserService
}

func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var user domain.User

	// แปลง JSON เป็น Struct
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// ส่งให้ Service (สมอง) จัดการ
	if err := h.service.Register(&user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// ลบ Password ออกจาก Response เพื่อความปลอดภัย
	user.Password = ""

	return c.Status(201).JSON(fiber.Map{
		"message": "สมัครสมาชิกสำเร็จ",
		"user":    user,
	})
}

// ==========================================
// 2. เข้าสู่ระบบ (Login)
// ==========================================
func (h *UserHandler) Login(c *fiber.Ctx) error {
	// สร้าง Struct เล็กๆ สำหรับรับค่า Login โดยเฉพาะ (DTO)
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	// เรียก Service ให้เช็ค Login และขอ Token
	token, role, err := h.service.Login(input.Username, input.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()}) // 401 Unauthorized
	}

	// ส่ง Token กลับไปให้ลูกค้า
	return c.JSON(fiber.Map{
		"message": "เข้าสู่ระบบสำเร็จ",
		"token":   token,
		"role":    role,
	})
}
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	idParam, _ := strconv.Atoi(c.Params("id"))

	requesterID := uint(c.Locals("user_id").(float64))
	requesterRole := c.Locals("role").(string)

	user, err := h.service.GetUser(requesterID, requesterRole, uint(idParam))
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	user.Password = ""
	return c.JSON(user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	idParam, _ := strconv.Atoi(c.Params("id"))

	requesterID := uint(c.Locals("user_id").(float64))
	requesterRole := c.Locals("role").(string)

	var input domain.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	err := h.service.UpdateUser(requesterID, requesterRole, uint(idParam), &input)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user updated"})
}
