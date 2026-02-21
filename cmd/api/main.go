package main

import (
	"log"

	"simple-clothes-shop/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment (dev)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 2. Create app
	application := app.NewApp()

	// 3. Run server
	application.Run()
}
