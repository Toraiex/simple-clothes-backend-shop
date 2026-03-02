package app

import (
	"context" // 1. ต้องเพิ่ม import context
	"log"
	"os"

	"simple-clothes-shop/internal/domain"
)

func SeedAdmin(userService domain.UserUsecase, userRepo domain.UserRepository) {

	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin123"
	}

	// เช็คก่อนว่ามีแล้วไหม
	existing, _ := userRepo.GetByUsername(context.Background(), username)
	if existing != nil {
		return
	}

	// ✅ 2. แก้ไขไวยากรณ์การสร้าง User และส่ง Context
	err := userService.Register(context.Background(), &domain.User{
		Username: username,
		Password: password,
		Role:     domain.RoleAdmin,
	})

	if err != nil {
		log.Println("🚨 seed admin failed:", err)
		return
	}

	log.Println("✅ admin created:", username)
}
