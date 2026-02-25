package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/repository"
	"simple-clothes-shop/internal/service"
	"simple-clothes-shop/pkg/database"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// --- Structs สำหรับรับข้อมูลจาก Platzi ---
type PlatziCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PlatziProduct struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Price       float64        `json:"price"`
	Description string         `json:"description"`
	Category    PlatziCategory `json:"category"`
	Images      []string       `json:"images"`
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

	// 📦 2. สร้าง Categories หลักแบบ Manual (ป้องกันข้อมูลขยะจาก API)
	seedCleanCategories(categoryService)

	// 👕 3. ดึงสินค้าจาก API แล้วจับคู่หมวดหมู่ให้ถูกต้อง
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
		err := svc.CreateCategory(name)
		if err != nil {
			log.Printf("❌ สร้างหมวดหมู่ '%s' ไม่สำเร็จ: %v\n", name, err)
		} else {
			fmt.Printf("   ✅ สร้างหมวดหมู่: %s\n", name)
		}
	}
}

func seedProductsAndVariants(db *sqlx.DB) {
	log.Println("...กำลังดึงข้อมูล Products จาก Platzi API")

	// ดึงสินค้ามา 40 ชิ้น
	resp, err := http.Get("https://api.escuelajs.co/api/v1/products?limit=40&offset=0")
	if err != nil {
		log.Fatalf("ดึงข้อมูล Products ไม่สำเร็จ: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("อ่านข้อมูล Body ไม่สำเร็จ: %v", err)
	}

	var platziProducts []PlatziProduct
	if err := json.Unmarshal(body, &platziProducts); err != nil {
		log.Fatalf("แปลง JSON ไม่สำเร็จ: %v", err)
	}

	colors := []string{"Red", "Blue", "Black", "White", "Green", "Grey"}
	sizes := []string{"S", "M", "L", "XL"}

	successCount := 0
	for _, p := range platziProducts {

		// 🧠 ระบบจับคู่หมวดหมู่อัจฉริยะ (Smart Category Mapping)
		// ไม่สนใจว่าจะสะกด Electronic หรือ Electronics เราจะจับยัดให้ถูกหมวด
		mappedCategory := "Miscellaneous" // ค่าเริ่มต้น
		lowerCatName := strings.ToLower(p.Category.Name)

		if strings.Contains(lowerCatName, "elect") {
			mappedCategory = "Electronics"
		} else if strings.Contains(lowerCatName, "cloth") || strings.Contains(lowerCatName, "shirt") {
			mappedCategory = "Clothes"
		} else if strings.Contains(lowerCatName, "shoe") {
			mappedCategory = "Shoes"
		} else if strings.Contains(lowerCatName, "furni") {
			mappedCategory = "Furniture"
		}

		// ดึง ID หมวดหมู่จากฐานข้อมูล
		var localCategoryID uint
		err := db.Get(&localCategoryID, "SELECT id FROM categories WHERE name=$1", mappedCategory)
		if err != nil {
			log.Printf("⚠️ ข้ามสินค้า '%s' (เกิดข้อผิดพลาดในการดึง ID หมวดหมู่)\n", p.Title)
			continue
		}

		var newProductID uint
		randomStock := rand.Intn(100) + 10
		imagesJSON, _ := json.Marshal(p.Images)

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
				log.Printf("   ⚠️ สร้าง Variant (%s) ไม่สำเร็จ: %v\n", sku, vErr)
			}
		}

		successCount++
		fmt.Printf("✔️ ได้สินค้า: %s -> อยู่ในหมวด: %s\n", p.Title, mappedCategory)
	}

	log.Printf("🎉 สร้างสินค้าพร้อมขายทั้งหมด %d รายการ!\n", successCount)
}
