package main

import (
	"fmt"
	"log/slog"
	"os"

	"simple-clothes-shop/internal/app"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {

	fmt.Println("👉 1. เริ่มรันฟังก์ชัน main")

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}
	fmt.Println("👉 2. โหลด .env เสร็จแล้ว")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logrus.Info("Starting API server on port 3000 🚀")

	fmt.Println("👉 3. กำลังจะสร้าง App (เชื่อมต่อ DB/Redis)")
	application := app.NewApp()

	application.Run()
}
