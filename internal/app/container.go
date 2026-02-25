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

func NewHandlersContainer(db *sqlx.DB) *HandlersContainer {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// ✅ 1. สร้าง Cart ขึ้นมาก่อน เพื่อให้มีตัวแปร cartRepo เอาไปใช้ต่อ
	cartRepo := repository.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo)
	cartHandler := handler.NewCartHandler(cartService)

	// ✅ 2. สร้าง Order ตามมา (ลบ productRepo ออก และโยน cartRepo เข้าไปแทน)
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, cartRepo) // แก้ไขบรรทัดนี้
	orderHandler := handler.NewOrderHandler(orderService)

	// 2. Services
	// ✅ สร้าง sessionRepo แยกออกมาก่อน แล้วโยน db เข้าไป
	sessionRepo := repository.NewSessionRepository(db)
	emailService := service.NewEmailService()
	userService := service.NewUserService(userRepo, sessionRepo, emailService)

	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo, productRepo)

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
