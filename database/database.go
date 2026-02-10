package database

import (
	"log"
	"simple-clothes-shop/models" // ⚠️ อย่าลืมเปลี่ยนชื่อ module ให้ตรงกับ go.mod ของคุณ

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ประกาศตัวแปร Global ระดับ Package (ตัวใหญ่เพื่อให้คนอื่นเรียกใช้ได้)
var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=postgres password=1234 dbname=postgres port=5432 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // เปิด Log SQL ให้เห็นชัดๆ แบบ Pro
	})

	if err != nil {
		log.Fatal("🔥 Failed to connect to database: ", err)
	}

	log.Println("✅ Connected to Database successfully")

	// Auto Migrate ย้ายมาไว้ที่นี่
	DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductVariant{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
	)
	log.Println("✅ Database Migrated")
}
