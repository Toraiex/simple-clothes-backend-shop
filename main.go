package main

import (
	"simple-clothes-shop/database"
	"simple-clothes-shop/routes" // 👈 เรียกใช้ไฟล์ Routes

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. เชื่อมต่อ Database
	database.Connect()

	// 2. สร้าง App
	app := fiber.New()

	// 3. ใส่ Middleware กลาง
	app.Use(logger.New()) // ช่วยดู Log เวลาเทส
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	// 4. เรียกใช้ Routes ทั้งหมดที่เราจัดไว้
	routes.Setup(app)

	// 5. Start Server
	app.Listen(":3000")
}
