// internal/app/container.go

package app

import (
	"simple-clothes-shop/internal/handler"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"

	"gorm.io/gorm"
)

// HandlersContainer ใช้เก็บ Handler ทั้งหมดที่จะส่งไปที่ Route
type HandlersContainer struct {
	User     *handler.UserHandler
	Product  *handler.ProductHandler
	Category *handler.CategoryHandler
}

// NewHandlersContainer ทำหน้าที่ Wiring ทุกอย่าง แล้วส่งคืนแค่ก้อน Handlers
func NewHandlersContainer(db *gorm.DB) *HandlersContainer {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// 2. Services
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	categoryService := service.NewCategoryService(categoryRepo)

	// 3. Return Handlers wrapped in a struct
	return &HandlersContainer{
		User:     handler.NewUserHandler(userService),
		Product:  handler.NewProductHandler(productService),
		Category: handler.NewCategoryHandler(categoryService),
	}
}
