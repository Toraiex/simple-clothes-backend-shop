package routes

import (
	"simple-clothes-shop/handlers" // 👈 เรียกใช้ Handlers

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	// 🟢 Public Routes
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)

	// (เดี๋ยวเราค่อยทยอยย้าย Product, Order มาใส่เพิ่มทีหลัง)

	// 🔒 Private Routes (ต้องมี Token)
	api := app.Group("/api", jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte("mysecretkey")},
	}))

	// สร้าง Route หลอกๆ ไว้เทสก่อน
	api.Get("/check", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "เข้าใช้งาน Private Zone ได้แล้ว!"})
	})
}
