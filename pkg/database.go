package database

import (
	"fmt"
	"log"
	"os" // สำหรับดึงค่า Environment
	"simple-clothes-shop/internal/domain"

	// เพิ่มตัวนี้

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
		&domain.User{},
		&domain.Category{},
		&domain.Product{},
		&domain.ProductVariant{},
	)
}
