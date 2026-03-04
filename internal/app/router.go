package app

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	// 💡 1. Import Infrastructure / Middleware
	cacheRepo "simple-clothes-shop/internal/cache/repository/redis" // สำหรับ CacheRepo
	"simple-clothes-shop/internal/mail"                             // สมมติว่าพี่ย้ายระบบอีเมลมาไว้ในนี้
	"simple-clothes-shop/internal/middleware"                       // สมมติว่าพี่ย้าย AuthMiddleware ไปไว้ในนี้

	// 💡 2. Import แบบ Alias (ตั้งชื่อย่อไม่ให้ตีกัน) ของแต่ละฟีเจอร์
	userHttp "simple-clothes-shop/internal/user/delivery/http"
	userPostgres "simple-clothes-shop/internal/user/repository/postgres"
	userUsecase "simple-clothes-shop/internal/user/usecase"

	categoryHttp "simple-clothes-shop/internal/category/delivery/http"
	categoryPostgres "simple-clothes-shop/internal/category/repository/postgres"
	categoryUsecase "simple-clothes-shop/internal/category/usecase"

	productHttp "simple-clothes-shop/internal/product/delivery/http"
	productPostgres "simple-clothes-shop/internal/product/repository/postgres"
	productUsecase "simple-clothes-shop/internal/product/usecase"

	cartHttp "simple-clothes-shop/internal/cart/delivery/http"
	cartPostgres "simple-clothes-shop/internal/cart/repository/postgres"
	cartUsecase "simple-clothes-shop/internal/cart/usecase"

	orderHttp "simple-clothes-shop/internal/order/delivery/http"
	orderPostgres "simple-clothes-shop/internal/order/repository/postgres"
	orderUsecase "simple-clothes-shop/internal/order/usecase"
)

func setupRoutes(app *fiber.App, db *sqlx.DB, rdb *redis.Client) {
	// กลุ่มเส้นทางหลัก
	api := app.Group("/api")

	// ==========================================
	// ⚙️ 1. Setup Infrastructure & Middlewares
	// ==========================================
	redisCache := cacheRepo.NewCacheRepository(rdb)
	emailSvc := mail.NewSMTPMailService() // สร้างบริการส่งอีเมล

	// สร้างยาม (Middlewares)
	authMid := middleware.NewAuthMiddleware(redisCache, os.Getenv("JWT_ACCESS_SECRET"))
	adminMid := middleware.IsAdmin()
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "ทำรายการบ่อยเกินไป กรุณารอสักครู่"})
		},
	})

	// ==========================================
	// 🚀 2. ประกอบร่างทีละ Feature (Dependency Injection)
	// ==========================================

	// --- 👤 Feature: User ---
	userRepo := userPostgres.NewUserRepository(db)
	userUC := userUsecase.NewUserService(userRepo, redisCache, emailSvc)
	userHttp.NewUserHandler(api, userUC, authMid, adminMid, authLimiter) // โยน API Router ให้มันผูก Route เอง

	// สร้าง Admin อัตโนมัติ (ถ้าระบบยังไม่มี)
	SeedAdmin(userUC, userRepo)

	// --- 🏷️ Feature: Category ---
	catRepo := categoryPostgres.NewCategoryRepository(db)
	// สมมติว่า CategoryUsecase ต้องใช้ ProductRepo ด้วย เราต้องสร้าง ProductRepo ออกมาก่อน
	prodRepo := productPostgres.NewProductRepository(db)
	catUC := categoryUsecase.NewCategoryUsecase(catRepo, prodRepo)
	categoryHttp.NewCategoryHandler(api, catUC, authMid, adminMid)

	// --- 👕 Feature: Product ---
	prodUC := productUsecase.NewProductUsecase(prodRepo, catRepo)
	productHttp.NewProductHandler(api, prodUC, authMid, adminMid)

	// --- 🛒 Feature: Cart ---
	cartRepo := cartPostgres.NewCartRepository(db)
	cartUC := cartUsecase.NewCartUsecase(cartRepo)
	cartHttp.NewCartHandler(api, cartUC, authMid)

	// --- 📦 Feature: Order ---
	orderRepo := orderPostgres.NewOrderRepository(db)
	orderUC := orderUsecase.NewOrderUsecase(orderRepo, cartRepo)
	orderHttp.NewOrderHandler(api, orderUC, authMid, adminMid)

}
