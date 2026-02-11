package database

import (
	"fmt"
	"log"
	"os" // สำหรับดึงค่า Environment
	"simple-clothes-shop/internal/domain"

	"github.com/joho/godotenv" // เพิ่มตัวนี้

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ประกาศตัวแปร Global ระดับ Package (ตัวใหญ่เพื่อให้คนอื่นเรียกใช้ได้)
var DB *gorm.DB

func Connect() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. ดึงค่าจาก .env มาประกอบเป็น DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // เปิด Log SQL ให้เห็นชัดๆ แบบ Pro
	})

	if err != nil {
		log.Fatal("🔥 Failed to connect to database: ", err)
	}

	log.Println("✅ Connected to Database successfully")

	// Auto Migrate ย้ายมาไว้ที่นี่
	DB.AutoMigrate(
		&domain.User{}, // 👈 เปลี่ยนจาก models.User เป็น domain.User
		&domain.Category{},
		&domain.Product{},
		&domain.ProductVariant{},
	)
	log.Println("✅ Database Migrated")
}
