package response

import "github.com/gofiber/fiber/v2"

// SuccessResponse รูปแบบข้อมูลเมื่อทำรายการสำเร็จ
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse รูปแบบข้อมูลเมื่อเกิดข้อผิดพลาด
func Error(c *fiber.Ctx, status int, message string, errDetail string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
		"error":   errDetail,
	})
}
