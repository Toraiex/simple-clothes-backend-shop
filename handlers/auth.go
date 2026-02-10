package handlers

import (
	"simple-clothes-shop/database" // 👈 เรียกใช้ Database ที่เราแยกไว้
	"simple-clothes-shop/models"   // 👈 เรียกใช้ Models ที่เราแยกไว้
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// DTO สำหรับรับค่า (ไม่ต้อง Export ก็ได้ถ้าใช้แค่ในนี้)
type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Role     string `json:"role"`
}

func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)

	// 🌟 เรียกใช้ models.User
	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Address:  input.Address,
		Role:     "user", // Default
	}

	if input.Role != "" {
		user.Role = input.Role
	}

	// 🌟 เรียกใช้ database.DB
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "สมัครสมาชิกไม่สำเร็จ (ชื่อซ้ำหรือระบบมีปัญหา)"})
	}

	return c.Status(201).JSON(fiber.Map{"message": "สมัครสมาชิกสำเร็จ", "user": user})
}

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	var user models.User
	// 🌟 ใช้ database.DB หา User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบผู้ใช้"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "รหัสผ่านผิด"})
	}

	// สร้าง JWT Token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("mysecretkey")) // ⚠️ ของจริงควรดึงจาก Env
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t, "role": user.Role})
}
