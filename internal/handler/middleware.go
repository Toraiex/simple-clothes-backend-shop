package handler

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 1. ยามตรวจบัตร (Check Token & Extract Role)
func AuthMiddleware(c *fiber.Ctx) error {
	// ✅ 1. ล้วงหาบัตร (Access Token) จากกระเป๋า (Cookie) แทนการดึงจาก Header
	tokenString := c.Cookies("access_token")

	// ถ้าไม่มีคุกกี้ส่งมา แสดงว่ายังไม่ได้ล็อกอิน หรือคุกกี้หมดอายุไปแล้ว (เกิน 15 นาที)
	if tokenString == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized: ไม่พบข้อมูลการเข้าสู่ระบบ หรือเซสชันหมดอายุ",
		})
	}

	// ✅ 2. ตรวจสอบความถูกต้องของ Token (ส่วนนี้ใช้ลอจิกเดิมของคุณได้เลย)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_ACCESS_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: token ไม่ถูกต้อง"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: invalid claims"})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: invalid user id"})
	}

	userID := uint(userIDFloat)
	role, ok := claims["role"].(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized: invalid role"})
	}

	// ✅ 3. ฝากข้อมูลไว้กับ Context (เหมือนเดิมเป๊ะ)
	c.Locals("user_id", userID)
	c.Locals("role", role)

	return c.Next()
}

func IsAdmin(c *fiber.Ctx) error {
	role, err := GetUserRole(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	if role != "admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "forbidden: admin only",
		})
	}

	return c.Next()
}
