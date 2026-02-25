package app

import (
	database "simple-clothes-shop/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type App struct {
	fiber *fiber.App
}

func NewApp() *App {
	// 1. Connect DB
	database.Connect()
	db := database.DB

	handlers := NewHandlersContainer(db)
	app := fiber.New()

	app.Use(
		requestid.New(), // สร้าง ID ให้ทุก Request
		logger.New(logger.Config{
			Format:     "${time} | ${pid} | ${locals:requestid} | ${status} | ${latency} | ${method} | ${path}\n",
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "Asia/Bangkok",
		}),
		recover.New(),
	)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server Running")
	})

	setupRoutes(app, handlers)

	return &App{fiber: app}
}

func (a *App) Run() {
	a.fiber.Listen(":3000")
}
