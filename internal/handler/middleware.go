package handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 1. ยามตรวจบัตร (Check Token & Extract Role)
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(401).JSON(fiber.Map{"error": "รูปแบบ Token ไม่ถูกต้อง (ต้องเป็น Bearer <token>)"})
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		fmt.Println("JWT ERROR:", err)
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	c.Locals("user_id", claims["user_id"])
	c.Locals("role", claims["role"])

	return c.Next()
}

// 2. ยามคัดกรองเฉพาะผู้บริหาร (Admin Only)
func IsAdmin(c *fiber.Ctx) error {
	// ดึงค่า role ที่เราฝากไว้ใน Locals มาดู
	role := c.Locals("role")

	if role != "admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "สิทธิ์ไม่เพียงพอ (เฉพาะ Admin เท่านั้น)",
		})
	}

	return c.Next()
}
