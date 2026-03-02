// package app

// import (
// 	"simple-clothes-shop/internal/handler"
// 	"simple-clothes-shop/internal/repository"
// 	database "simple-clothes-shop/pkg/database"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/logger"
// 	"github.com/gofiber/fiber/v2/middleware/recover"
// 	"github.com/gofiber/fiber/v2/middleware/requestid"
// 	"github.com/redis/go-redis/v9"
// )

// type App struct {
// 	fiber *fiber.App
// }

// func NewApp() *App {
// 	// 1. Connect DB
// 	database.Connect()
// 	db := database.DB
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     "127.0.0.1:6379", // ปกติ Redis รันที่พอร์ตนี้
// 		Password: "",               // ปล่อยว่างถ้าไม่ได้ตั้งรหัส
// 		DB:       0,                // ใช้ DB 0
// 	})
// 	cacheRepo := repository.NewCacheRepository(rdb)
// 	handlers := NewHandlersContainer(db, rdb)
// 	authMid := handler.NewAuthMiddleware(cacheRepo)
// 	adminMid := handler.IsAdmin()
// 	app := fiber.New()

// 	app.Use(
// 		requestid.New(), // สร้าง ID ให้ทุก Request
// 		logger.New(logger.Config{
// 			Format:     "${time} | ${pid} | ${locals:requestid} | ${status} | ${latency} | ${method} | ${path}\n",
// 			TimeFormat: "2006-01-02 15:04:05",
// 			TimeZone:   "Asia/Bangkok",
// 		}),
// 		recover.New(),
// 	)
// 	app.Get("/", func(c *fiber.Ctx) error {
// 		return c.SendString("Server Running")
// 	})

// 	setupRoutes(app, handlers, authMid, adminMid)

// 	return &App{fiber: app}
// }

//	func (a *App) Run() {
//		a.fiber.Listen(":3000")
//	}
package app

import (
	database "simple-clothes-shop/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/redis/go-redis/v9"
)

type App struct {
	fiber *fiber.App
}

func NewApp() *App {
	// 1. เชื่อมต่อฐานข้อมูลหลัก
	database.Connect()
	db := database.DB

	// 2. เชื่อมต่อ Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	app := fiber.New()

	// 3. ใส่ Global Middlewares
	app.Use(
		requestid.New(),
		logger.New(logger.Config{
			Format:     "${time} | ${status} | ${latency} | ${method} | ${path}\n",
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "Asia/Bangkok",
		}),
		recover.New(),
	)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Simple Clothes API is running! 🚀")
	})

	// 🚀 4. โยน db และ rdb ไปประกอบร่างที่ router.go (ไม่ต้องพึ่ง container แล้ว)
	setupRoutes(app, db, rdb)

	return &App{fiber: app}
}

func (a *App) Run() {
	a.fiber.Listen(":3000")
}
