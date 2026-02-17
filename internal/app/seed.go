package app

import (
	"log"
	"os"

	"simple-clothes-shop/internal/domain"
)

func SeedAdmin(userService domain.UserService, userRepo domain.UserRepository) {

	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin123"
	}

	// เช็คก่อนว่ามีแล้วไหม
	existing, _ := userRepo.GetByUsername(username)
	if existing != nil {
		return
	}

	// ถ้าไม่มี → สร้าง
	err := userService.Register(&domain.User{
		Username: username,
		Password: password,
		Role:     domain.RoleAdmin,
	})

	if err != nil {
		log.Println("seed admin failed:", err)
		return
	}

	log.Println("✅ admin created:", username)
}
