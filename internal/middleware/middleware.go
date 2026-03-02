package middleware

import (
	"os"
	// 👈 อย่าลืมเช็ค path ให้ตรงกับโปรเจกต์คุณนะครับ

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 1. ยามตรวจบัตร (อัปเกรด: รับตัวเชื่อมต่อ Redis เข้ามาด้วย)
// เปลี่ยนจาก func ธรรมดา เป็น func ที่ return fiber.Handler ตามมาตรฐาน Fiber
func NewAuthMiddleware(cacheRepo redisRepo.CacheRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ✅ 1. ล้วงหาบัตรทั้ง 2 ใบ จากกระเป๋า (Cookie)
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")

		// ถ้าไม่มีคุกกี้ใบใดใบหนึ่ง แสดงว่ายังไม่ได้ล็อกอิน หรือหลุดออกจากระบบแล้ว
		if accessToken == "" || refreshToken == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: ไม่พบข้อมูลการเข้าสู่ระบบ หรือเซสชันหมดอายุ",
			})
		}

		// 🔥 2. ด่านสกัด Session ผี: ถาม Redis ว่า Refresh Token นี้ยังอยู่ไหม?
		// ส่ง c.UserContext() เข้าไปด้วยตามมาตรฐานใหม่ของเรา
		_, err := cacheRepo.GetSession(c.UserContext(), refreshToken)
		if err != nil {
			// ถ้าหาใน Redis ไม่เจอ = เพิ่งกด Logout หรือโดน Admin เตะออก
			// สั่งลบ Cookie ฝั่งหน้าเว็บทิ้งไปเลยเพื่อความสะอาด
			c.ClearCookie("access_token", "refresh_token")
			return c.Status(401).JSON(fiber.Map{
				"error": "unauthorized: เซสชันถูกยกเลิก กรุณาล็อกอินใหม่",
			})
		}

		// ✅ 3. ตรวจสอบความถูกต้องของ Access Token (คณิตศาสตร์)
		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
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

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized: invalid role"})
		}

		// ✅ 4. ฝากข้อมูลไว้กับ Context เพื่อให้ Handler ตัวอื่นดึงไปใช้ต่อได้
		c.Locals("user_id", uint(userIDFloat))
		c.Locals("role", role)

		return c.Next()
	}
}

// 2. ด่านตรวจสิทธิ์ Admin (Refactor ให้คลีนขึ้น ไม่ต้องพึ่ง GetUserRole)
func IsAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึง Role จากที่ AuthMiddleware ฝากไว้ใน Locals ได้เลย
		role, ok := c.Locals("role").(string)

		if !ok || role != "admin" {
			return c.Status(403).JSON(fiber.Map{
				"error": "forbidden: สิทธิ์การเข้าถึงถูกปฏิเสธ เฉพาะผู้ดูแลระบบเท่านั้น",
			})
		}

		return c.Next()
	}
}
