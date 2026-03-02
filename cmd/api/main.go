package main

import (
	"log"
	"log/slog"
	"os"

	"simple-clothes-shop/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment (dev)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // แสดง Log ตั้งแต่ระดับ Info ขึ้นไป (Info, Warn, Error)
	}))
	slog.SetDefault(logger) // บังคับให้ทั้งโปรเจกต์ใช้ Logger ตัวนี้เป็นค่าเริ่มต้น

	slog.Info("Starting API server", slog.String("port", "3000"))

	// 2. Create app
	application := app.NewApp()

	// 3. Run server
	application.Run()
}
