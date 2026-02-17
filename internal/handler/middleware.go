package handler

import (
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
		return c.Status(401).JSON(fiber.Map{
			"error": "รูปแบบ Token ไม่ถูกต้อง",
		})
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid claims"})
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid user id"})
	}

	userID := uint(userIDFloat)

	role, ok := claims["role"].(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid role"})
	}

	SetUserContext(c, userID, role)

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
