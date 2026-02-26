// internal/app/container.go

package app

import (
	"os"
	"simple-clothes-shop/internal/handler"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// HandlersContainer ใช้เก็บ Handler ทั้งหมดที่จะส่งไปที่ Route
type HandlersContainer struct {
	User     *handler.UserHandler
	Product  *handler.ProductHandler
	Category *handler.CategoryHandler
	Order    *handler.OrderHandler
	Cart     *handler.CartHandler
}

func NewHandlersContainer(db *sqlx.DB, rdb *redis.Client) *HandlersContainer {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	cacheRepo := repository.NewCacheRepository(rdb) // ✅ ตอนนี้รู้จัก rdb แล้ว
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// 2. Services
	cartService := service.NewCartService(cartRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo)

	emailService := service.NewEmailService()
	userService := service.NewUserService(userRepo, cacheRepo, emailService)
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo, productRepo)

	// 3. Handlers
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)

	if os.Getenv("AUTO_SEED_ADMIN") == "true" {
		SeedAdmin(userService, userRepo)
	}

	// 4. Return Handlers wrapped in a struct
	return &HandlersContainer{
		User:     handler.NewUserHandler(userService),
		Product:  handler.NewProductHandler(productService),
		Category: handler.NewCategoryHandler(categoryService),
		Order:    orderHandler,
		Cart:     cartHandler,
	}
}
