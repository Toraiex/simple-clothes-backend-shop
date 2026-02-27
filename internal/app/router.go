package app

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// 🚀 เพิ่ม authMid และ adminMid เข้ามาเป็นพารามิเตอร์รับค่าจาก main.go
func setupRoutes(app *fiber.App, h *HandlersContainer, authMid fiber.Handler, adminMid fiber.Handler) {
	api := app.Group("/api")

	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "คุณทำรายการบ่อยเกินไป กรุณารอสักครู่แล้วลองใหม่",
			})
		},
	})

	// Auth
	api.Post("/register", h.User.Register)
	api.Post("/login", authLimiter, h.User.Login)
	api.Post("/refresh", h.User.RefreshToken)
	api.Post("/logout", h.User.Logout)

	api.Post("/verify-email", h.User.VerifyEmail)
	api.Post("/resend-otp", authLimiter, h.User.ResendOTP)

	api.Post("/forgot-password", h.User.ForgotPassword)
	api.Post("/resend-reset-otp", authLimiter, h.User.ResendResetOTP)
	api.Post("/reset-password", h.User.ResetPassword)

	// 🚀 เปลี่ยน handler.AuthMiddleware เป็น authMid และ handler.IsAdmin เป็น adminMid
	api.Get("/users", authMid, adminMid, h.User.GetAllUsers)
	api.Get("/users/:id", authMid, h.User.GetUser)
	api.Patch("/users/:id", authMid, h.User.UpdateUser)

	// Categories
	api.Post("/categories", authMid, adminMid, h.Category.Create)
	api.Get("/categories", h.Category.GetAll)
	api.Get("/categories/:id", h.Category.GetByID)
	api.Put("/categories/:id", authMid, adminMid, h.Category.Update)
	api.Delete("/categories/:id", authMid, adminMid, h.Category.Delete)

	// Products
	api.Get("/products", h.Product.GetAll)
	api.Get("/products/:id", h.Product.GetByID)
	api.Post("/products", authMid, adminMid, h.Product.Create)
	api.Patch("/products/:id", authMid, adminMid, h.Product.Update)
	api.Delete("/products/:id", authMid, adminMid, h.Product.Delete)
	api.Delete("/products/variants/:id", authMid, adminMid, h.Product.DeleteVariant)

	// Carts (ตะกร้าสินค้าของฉัน)
	api.Get("/cart", authMid, h.Cart.GetMyCart)
	api.Post("/cart", authMid, h.Cart.AddToCart)
	api.Patch("/cart/items/:id", authMid, h.Cart.UpdateQuantity)
	api.Delete("/cart/items/:id", authMid, h.Cart.RemoveFromCart)

	// Orders
	api.Post("/orders", authMid, h.Order.Checkout)
	api.Get("/orders/:id", authMid, h.Order.GetByID)
	api.Get("/orders", authMid, h.Order.GetMyOrders)
	api.Put("/orders/:id/cancel", authMid, h.Order.Cancel)
	api.Put("/orders/:id/status", authMid, adminMid, h.Order.AdminUpdateStatus)
}
