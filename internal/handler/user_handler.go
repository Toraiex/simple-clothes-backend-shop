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

// ... (ส่วนหน้าของไฟล์ และฟังก์ชัน Register, Login, GetUser เหมือนเดิม) ...

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
