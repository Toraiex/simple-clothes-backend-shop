// internal/app/container.go

package app

import (
	"os"
	"simple-clothes-shop/internal/handler"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"

	"github.com/jmoiron/sqlx"
)

// HandlersContainer ใช้เก็บ Handler ทั้งหมดที่จะส่งไปที่ Route
type HandlersContainer struct {
	User     *handler.UserHandler
	Product  *handler.ProductHandler
	Category *handler.CategoryHandler
	Order    *handler.OrderHandler
	Cart     *handler.CartHandler
}

// NewHandlersContainer ทำหน้าที่ Wiring ทุกอย่าง แล้วส่งคืนแค่ก้อน Handlers
func NewHandlersContainer(db *sqlx.DB) *HandlersContainer {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, productRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	cartRepo := repository.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo)
	cartHandler := handler.NewCartHandler(cartService)

	// 2. Services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	if os.Getenv("AUTO_SEED_ADMIN") == "true" {
		SeedAdmin(userService, userRepo)
	}

	// 3. Return Handlers wrapped in a struct
	return &HandlersContainer{
		User:     handler.NewUserHandler(userService),
		Product:  handler.NewProductHandler(productService),
		Category: handler.NewCategoryHandler(categoryService),
		Order:    orderHandler, // ✅ เพิ่ม\
		Cart:     cartHandler,
	}

}
