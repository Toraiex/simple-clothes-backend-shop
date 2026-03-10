package middleware

import (
	"simple-clothes-shop/internal/domain"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
)

func NewAuthLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,               // ยิงได้สูงสุด 5 ครั้ง
		Expiration: 1 * time.Minute, // ภายใน 1 นาที
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // บล็อกตาม IP Address ของแฮกเกอร์
		},
		LimitReached: func(c *fiber.Ctx) error {
			// 💡 ท่าไม้ตาย: ส่ง 429 พร้อมข้อความมาตรฐานของเรา!
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": domain.ErrTooManyRequests.Error(),
			})
		},
	})
}

func NewAuthMiddleware(cacheRepo domain.CacheRepository, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")

		if accessToken == "" || refreshToken == "" {
			// 💡 ใช้ domain.ErrUnauthorized.Error() แทนการพิมพ์สด
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": domain.ErrUnauthorized.Error(),
			})
		}

		// ด่านสกัด Session ผี
		_, err := cacheRepo.GetSession(c.UserContext(), refreshToken)
		if err != nil {
			c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1, HTTPOnly: true})
			c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", Path: "/", MaxAge: -1, HTTPOnly: true})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "เซสชันถูกยกเลิก หรือหมดอายุ กรุณาล็อกอินใหม่",
			})
		}

		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": domain.ErrUnauthorized.Error()})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": domain.ErrUnauthorized.Error()})
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": domain.ErrUnauthorized.Error()})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": domain.ErrUnauthorized.Error()})
		}

		// ใช้ Helper จากไฟล์ context.go ที่คุณเขียนไว้ได้เลยครับ!
		SetUserContext(c, uint(userIDFloat), role)

		return c.Next()
	}
}

// ==========================================================
// 👑 3. Admin Middleware (ตรวจสิทธิ์ผู้ดูแลระบบ)
// ==========================================================
func IsAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// เรียกใช้ Helper ของคุณให้เป็นประโยชน์
		role, err := GetUserRole(c)
		if err != nil || role != string(domain.RoleAdmin) {
			// 💡 ใช้ domain.ErrForbidden
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": domain.ErrForbidden.Error(),
			})
		}
		return c.Next()
	}
}
