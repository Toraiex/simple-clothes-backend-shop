package app

import (
	"simple-clothes-shop/internal/handler"

	"github.com/gofiber/fiber/v2"
)


func setupRoutes(app *fiber.App, h *HandlersContainer) {
	api := app.Group("/api")

	// Auth
	api.Post("/register", h.User.Register)
	api.Post("/login", h.User.Login)

	// Users
	api.Get("/users/:id", handler.AuthMiddleware, h.User.GetUser)
	api.Put("/users/:id", handler.AuthMiddleware, h.User.UpdateUser)

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
	api.Put("/products/:id", handler.AuthMiddleware, handler.IsAdmin, h.Product.Update)
	api.Delete("/products/:id", handler.AuthMiddleware, handler.IsAdmin, h.Product.Delete)

	api.Post("/orders", handler.AuthMiddleware, h.Order.Create)
	api.Get("/orders/:id", handler.AuthMiddleware, h.Order.GetByID)
	api.Get("/orders", handler.AuthMiddleware, h.Order.GetMyOrders)
	api.Put("/orders/:id/cancel", handler.AuthMiddleware, h.Order.Cancel)
	api.Put("/orders/:id/status", handler.AuthMiddleware, handler.IsAdmin, h.Order.AdminUpdateStatus)

}
