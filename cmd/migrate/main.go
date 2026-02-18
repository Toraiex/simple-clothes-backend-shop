package main

import (
	"flag"
	"log"

	"simple-clothes-shop/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// ✅ 1. สร้าง Flag ชื่อ action (ค่าเริ่มต้นคือ "up")
	action := flag.String("action", "up", "Migration action: up, down, drop")
	flag.Parse()

	// 2. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 3. Connect DB (มันจะไปอ่าน DB_HOST, DB_USER จาก .env ของคุณอัตโนมัติ)
	database.Connect()
	db := database.DB

	// 4. สั่งรัน Migration โดยส่งค่า action ไปด้วย
	log.Printf("🚀 Starting database migration with action: [%s]...", *action)

	database.RunMigrations(db.DB, *action)

	log.Println("✨ Migration process finished")
}
