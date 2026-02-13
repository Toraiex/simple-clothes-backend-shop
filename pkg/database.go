package database

import (
	"fmt"
	"log"
	"os" // สำหรับดึงค่า Environment

	"simple-clothes-shop/internal/repository"

	// เพิ่มตัวนี้v
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ประกาศตัวแปร Global ระดับ Package (ตัวใหญ่เพื่อให้คนอื่นเรียกใช้ได้)
var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("🔥 Failed to connect to database: ", err)
	}

	DB = db

}

func Migrate() {
	DB.AutoMigrate(
		&repository.CategoryModel{},
		&repository.ProductModel{},
		&repository.ProductVariantModel{},
		&repository.UserModel{},
		&repository.OrderModel{},
		&repository.OrderItemModel{},
	)
}
func SeedAdmin() {
	username := os.Getenv("SEED_ADMIN_USERNAME")
	password := os.Getenv("SEED_ADMIN_PASSWORD")

	// ถ้าไม่ได้ตั้ง ENV → ไม่ seed
	if username == "" || password == "" {
		fmt.Println("⚠️ Admin seed skipped (ENV not set)")
		return
	}

	var count int64

	// เช็คจาก username (ปลอดภัยกว่าเช็ค role)
	DB.Model(&repository.UserModel{}).
		Where("username = ?", username).
		Count(&count)

	if count > 0 {
		fmt.Println("ℹ️ Admin already exists, skipping seed")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Println("❌ Failed to hash admin password")
		return
	}

	err = DB.Create(&repository.UserModel{
		Username: username,
		Password: string(hashedPassword),
		Role:     "admin",
	}).Error

	if err != nil {
		fmt.Println("❌ Failed to create admin:", err)
		return
	}

	fmt.Println("✅ Admin account seeded successfully")
}
