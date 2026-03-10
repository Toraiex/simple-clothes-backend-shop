package main

import (
	"flag"
	"simple-clothes-shop/pkg/database"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus" // 💡 ใช้ logrus
)

func main() {
	action := flag.String("action", "up", "Migration action: up, down, drop")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	database.Connect()
	db := database.DB

	logrus.Infof("🚀 Starting database migration with action: [%s]...", *action)

	database.RunMigrations(db.DB, *action)

	logrus.Info("✨ Migration process finished successfully")
}
