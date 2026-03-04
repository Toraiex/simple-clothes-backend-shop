package middleware

import (
	"simple-clothes-shop/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 💉 Inject ทั้ง CacheRepo และ JWT Secret เข้ามาแต่แรก
func NewAuthMiddleware(cacheRepo domain.CacheRepository, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")

		if accessToken == "" || refreshToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized: ไม่พบข้อมูลการเข้าสู่ระบบ หรือเซสชันหมดอายุ",
			})
		}

		// 🔥 ด่านสกัด Session ผี
		_, err := cacheRepo.GetSession(c.UserContext(), refreshToken)
		if err != nil {
			// ระบุค่าให้ชัดเจนเพื่อบังคับลบ Cookie อย่างสมบูรณ์แบบ
			c.Cookie(&fiber.Cookie{
				Name:     "access_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HTTPOnly: true,
			})
			c.Cookie(&fiber.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HTTPOnly: true,
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized: เซสชันถูกยกเลิก กรุณาล็อกอินใหม่",
			})
		}

		// ✅ ตรวจสอบ Access Token (ใช้ Secret ที่ Inject เข้ามา)
		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil // 🚀 ใช้ตัวแปร แทนการเรียก os.Getenv
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized: token ไม่ถูกต้อง"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized: invalid claims"})
		}

		// Parse ข้อมูล
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized: invalid user id"})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized: invalid role"})
		}

		// ฝาก Context
		c.Locals("user_id", uint(userIDFloat))
		c.Locals("role", role)

		return c.Next()
	}
}

func IsAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != string(domain.RoleAdmin) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden: admin only",
			})
		}
		return c.Next()
	}
}
