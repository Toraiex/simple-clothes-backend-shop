package app

import (
	database "simple-clothes-shop/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	// 4. Middleware
	app.Use(
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PATCH,DELETE,PUT",
		}),
		logger.New(logger.Config{
			TimeFormat: "2006-01-02 15:04:05",
			TimeZone:   "Asia/Bangkok",
		}),
	)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server Running")
	})

	// 5. Routes
	setupRoutes(app, handlers)

	return &App{fiber: app}
}

func (a *App) Run() {
	a.fiber.Listen(":3000")
}
