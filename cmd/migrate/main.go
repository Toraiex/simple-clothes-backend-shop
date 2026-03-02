package main

import (
	"flag"
	"log"

	"simple-clothes-shop/pkg/database"

	"github.com/joho/godotenv"
)

func main() {

	action := flag.String("action", "up", "Migration action: up, down, drop")
	flag.Parse()

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
