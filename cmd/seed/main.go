package main

import (
	"context"
	"encoding/json" // 👈 ต้องใช้สำหรับแปลง JSON
	"fmt"
	"io" // 👈 ต้องใช้สำหรับอ่านข้อมูลจาก API
	"log"
	"math/rand"
	"net/http" // 👈 ต้องใช้สำหรับยิง API
	"strings"
	"time"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"
	"simple-clothes-shop/pkg/database"

	"github.com/jmoiron/sqlx" // 👈 ต้องใช้สำหรับ *sqlx.DB
	"github.com/joho/godotenv"
)

// --- Structs สำหรับรับข้อมูลจาก FakeStore API ---
type FakeStoreProduct struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	rand.Seed(time.Now().UnixNano())

	database.Connect()
	db := database.DB

	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo, productRepo)

	log.Println("🌱 กำลังเริ่มกระบวนการ Seed ข้อมูล...")

	// 🧹 1. ล้างข้อมูลเก่าทิ้งทั้งหมด (Reset Database)
	log.Println("...กำลังทำความสะอาดตาราง (Truncate)")
	_, _ = db.Exec("TRUNCATE TABLE categories, products, product_variants, cart_items, carts, order_items, orders RESTART IDENTITY CASCADE")

	// 📦 2. สร้าง Categories หลักแบบ Manual
	seedCleanCategories(categoryService)

	// 👕 3. ดึงสินค้าจาก API
	seedProductsAndVariants(db)

	log.Println("✅ กระบวนการ Seed ข้อมูลเสร็จสมบูรณ์!")
}

func seedCleanCategories(svc domain.CategoryService) {
	log.Println("...กำลังสร้างหมวดหมู่หลัก (Core Categories)")

	coreCategories := []string{
		"Clothes",
		"Electronics",
		"Furniture",
		"Shoes",
		"Miscellaneous",
	}

	for _, name := range coreCategories {
		err := svc.CreateCategory(context.Background(), name)
		if err != nil {
			log.Printf("❌ สร้างหมวดหมู่ '%s' ไม่สำเร็จ: %v\n", name, err)
		} else {
			fmt.Printf("   ✅ สร้างหมวดหมู่: %s\n", name)
		}
	}
}

// 👇 นี่คือฟังก์ชันที่หายไปจากไฟล์ของคุณครับ! เอามาต่อไว้ล่างสุดแล้ว
func seedProductsAndVariants(db *sqlx.DB) {
	log.Println("...กำลังดึงข้อมูลจาก FakeStore API (Stable Version)")

	resp, err := http.Get("https://fakestoreapi.com/products")
	if err != nil {
		log.Fatalf("ดึงข้อมูล Products ไม่สำเร็จ: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("อ่านข้อมูล Body ไม่สำเร็จ: %v", err)
	}

	var fakeProducts []FakeStoreProduct
	if err := json.Unmarshal(body, &fakeProducts); err != nil {
		log.Fatalf("แปลง JSON ไม่สำเร็จ: %v", err)
	}

	log.Printf("📥 โหลดสินค้ามาได้ %d ชิ้น... เริ่มนำเข้าฐานข้อมูล\n", len(fakeProducts))

	colors := []string{"Red", "Blue", "Black", "White", "Green", "Grey"}
	sizes := []string{"S", "M", "L", "XL"}

	successCount := 0
	for _, p := range fakeProducts {
		mappedCategory := "Miscellaneous"
		if p.Category == "men's clothing" || p.Category == "women's clothing" {
			mappedCategory = "Clothes"
		} else if p.Category == "electronics" {
			mappedCategory = "Electronics"
		} else if p.Category == "jewelery" {
			mappedCategory = "Miscellaneous"
		}

		var localCategoryID uint
		err := db.Get(&localCategoryID, "SELECT id FROM categories WHERE name=$1", mappedCategory)
		if err != nil {
			continue
		}

		var newProductID uint
		randomStock := rand.Intn(100) + 10
		imagesJSON, _ := json.Marshal([]string{p.Image})

		err = db.QueryRow(`
			INSERT INTO products (name, description, price, stock, category_id, images)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, p.Title, p.Description, p.Price, randomStock, localCategoryID, imagesJSON).Scan(&newProductID)

		if err != nil {
			log.Printf("❌ สร้างสินค้า '%s' ไม่สำเร็จ: %v\n", p.Title, err)
			continue
		}

		numVariants := rand.Intn(3) + 2
		for i := 0; i < numVariants; i++ {
			color := colors[rand.Intn(len(colors))]
			size := sizes[rand.Intn(len(sizes))]
			variantStock := rand.Intn(20) + 1
			variantPrice := p.Price

			sku := fmt.Sprintf("PRD-%d-%s-%s-%d", newProductID, strings.ToUpper(color), size, i)

			attributesMap := map[string]string{
				"Color": color,
				"Size":  size,
			}
			attributesJSON, _ := json.Marshal(attributesMap)

			_, vErr := db.Exec(`
				INSERT INTO product_variants (product_id, sku, attributes, price, stock) 
				VALUES ($1, $2, $3, $4, $5)
			`, newProductID, sku, attributesJSON, variantPrice, variantStock)

			if vErr != nil {
				// ข้ามเงียบๆ ถ้า error
			}
		}

		successCount++
		fmt.Printf("✔️ ได้สินค้า: %s -> อยู่ในหมวด: %s\n", p.Title, mappedCategory)
	}

	log.Printf("🎉 สร้างสินค้าพร้อมขายทั้งหมด %d รายการ!\n", successCount)
}
