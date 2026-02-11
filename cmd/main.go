package main

import (
	"fmt"
	"simple-clothes-shop/internal/handler"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"
	database "simple-clothes-shop/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. เชื่อมต่อฐานข้อมูล

	database.Connect()
	db := database.DB

	// ------------------------------------------------
	// ⚡ ประกอบร่าง (Wiring)
	// ------------------------------------------------

	// User Component
	// 1. สร้างพวก Repository ก่อนให้หมด
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db) // 👈 สร้างก่อนเพื่อน

	// 2. สร้าง Service (ส่ง Repo เข้าไป)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	// 👈 ตอนนี้เรียกใช้ categoryRepo ได้แล้ว และห้ามใส่ := ซ้ำ
	productService := service.NewProductService(productRepo, categoryRepo)

	// 3. สร้าง Handler
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)

	// ------------------------------------------------
	// 🚀 ตั้งค่า Server & Middleware
	// ------------------------------------------------
	app := fiber.New()
	app.Use(

		cors.New(cors.Config{

			AllowOrigins: "*",

			AllowMethods: "GET,POST,PATCH,DELETE,PUT",
		}),

		logger.New(logger.Config{

			TimeFormat: "2006-01-02 15:04:05",

			TimeZone: "Asia/Bangkok",
		}),
	)

	// ------------------------------------------------
	// 📍 กำหนดเส้นทาง (Routes) - ทำหน้าที่แทน routes.go
	// ------------------------------------------------
	api := app.Group("/api")

	// Auth Routes
	api.Post("/register", userHandler.Register) // เรียกผ่านตัวแปร userHandler ที่ประกอบร่างแล้ว
	api.Post("/login", userHandler.Login)

	api.Get("/users/:id", handler.AuthMiddleware, userHandler.GetUser)
	api.Put("/users/:id", handler.AuthMiddleware, userHandler.UpdateUser)

	api.Post("/categories", handler.AuthMiddleware, handler.IsAdmin, categoryHandler.Create)
	api.Get("/categories", categoryHandler.GetAll)      // ดูทั้งหมด
	api.Get("/categories/:id", categoryHandler.GetByID) // ดูหมวดเดียวพร้อมสินค้าในนั้น
	api.Put("/categories/:id", handler.AuthMiddleware, handler.IsAdmin, categoryHandler.Update)
	api.Delete("/categories/:id", handler.AuthMiddleware, handler.IsAdmin, categoryHandler.Delete)

	// Product Routes
	api.Get("/products", productHandler.GetAll)
	api.Get("/products/:id", productHandler.GetByID)
	api.Put("/products/:id", handler.AuthMiddleware, handler.IsAdmin, productHandler.Update)

	// (เดี๋ยวเราค่อยมาใส่ Middleware กันตรงนี้ทีหลัง)
	api.Post("/products", handler.AuthMiddleware, handler.IsAdmin, productHandler.Create)
	api.Delete("/products/:id", handler.AuthMiddleware, handler.IsAdmin, productHandler.Delete)

	// Start Server

	for _, r := range app.GetRoutes() {
		fmt.Println(r.Method, r.Path)
	}
	app.All("*", func(c *fiber.Ctx) error {
		fmt.Println("HIT ROUTE:", c.Method(), c.Path())
		return c.Next()
	})

	app.Listen(":3000")
}
