package app

import (
	"simple-clothes-shop/internal/handler"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func setupRoutes(app *fiber.App, h *HandlersContainer) {
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

	api.Get("/users", handler.AuthMiddleware, handler.IsAdmin, h.User.GetAllUsers) // 👈 เพิ่มใหม่สำหรับ Admin
	api.Get("/users/:id", handler.AuthMiddleware, h.User.GetUser)
	api.Patch("/users/:id", handler.AuthMiddleware, h.User.UpdateUser) // 👈 เปลี่ยนจาก Put เป็น Patch ให้ถูกต้องตามหลั

	// Categories
	api.Post("/categories", handler.AuthMiddleware, handler.IsAdmin, h.Category.Create)
	api.Get("/categories", h.Category.GetAll)
	api.Get("/categories/:id", h.Category.GetByID)
	api.Put("/categories/:id", handler.AuthMiddleware, handler.IsAdmin, h.Category.Update)
	api.Delete("/categories/:id", handler.AuthMiddleware, handler.IsAdmin, h.Category.Delete)

	// Products
	api.Get("/products", h.Product.GetAll)
	api.Get("/products/:id", h.Product.GetByID)
	api.Post("/products", handler.AuthMiddleware, handler.IsAdmin, h.Product.Create)
	// เปลี่ยนบรรทัดนี้
	api.Patch("/products/:id", handler.AuthMiddleware, handler.IsAdmin, h.Product.Update)
	api.Delete("/products/:id", handler.AuthMiddleware, handler.IsAdmin, h.Product.Delete)
	api.Delete("/products/variants/:id", handler.AuthMiddleware, handler.IsAdmin, h.Product.DeleteVariant)

	// Carts (ตะกร้าสินค้าของฉัน)
	api.Get("/cart", handler.AuthMiddleware, h.Cart.GetMyCart)                   // ดูตะกร้า
	api.Post("/cart", handler.AuthMiddleware, h.Cart.AddToCart)                  // หยิบของใส่ตะกร้า
	api.Patch("/cart/items/:id", handler.AuthMiddleware, h.Cart.UpdateQuantity)  // แก้ไขจำนวนชิ้น
	api.Delete("/cart/items/:id", handler.AuthMiddleware, h.Cart.RemoveFromCart) // ลบของทิ้ง

	api.Post("/orders", handler.AuthMiddleware, h.Order.Checkout) // 👈 เปลี่ยนชื่อฟังก์ชันตรงนี้
	api.Get("/orders/:id", handler.AuthMiddleware, h.Order.GetByID)
	api.Get("/orders", handler.AuthMiddleware, h.Order.GetMyOrders)
	api.Put("/orders/:id/cancel", handler.AuthMiddleware, h.Order.Cancel)
	api.Put("/orders/:id/status", handler.AuthMiddleware, handler.IsAdmin, h.Order.AdminUpdateStatus)

}
