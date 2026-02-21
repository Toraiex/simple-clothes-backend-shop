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

	// ตั้งค่า Seed สำหรับการสุ่ม
	rand.Seed(time.Now().UnixNano())

	database.Connect()
	db := database.DB

	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)

	log.Println("🌱 กำลังเริ่มกระบวนการ Seed ข้อมูลจาก Platzi...")

	// 1. สร้าง Categories ก่อน
	seedCategories(categoryService)

	// 2. สร้าง Products และ Variants ต่อเลย
	seedProductsAndVariants(db)

	log.Println("✅ กระบวนการ Seed ข้อมูลเสร็จสมบูรณ์!")
}

func seedCategories(svc domain.CategoryService) {
	log.Println("...กำลังดึงข้อมูล Categories")
	resp, err := http.Get("https://api.escuelajs.co/api/v1/categories")
	if err != nil {
		log.Fatalf("ดึงข้อมูลไม่สำเร็จ: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("อ่านข้อมูล Body ไม่สำเร็จ: %v", err)
	}

	var platziCategories []PlatziCategory
	_ = json.Unmarshal(body, &platziCategories)

	for _, cat := range platziCategories {
		err := svc.CreateCategory(cat.Name)
		if err != nil && !strings.Contains(err.Error(), "มีอยู่ในระบบแล้ว") {
			log.Printf("❌ สร้างหมวดหมู่ '%s' ไม่สำเร็จ: %v\n", cat.Name, err)
		}
	}
	log.Println("✔️ จัดการข้อมูลหมวดหมู่เสร็จสิ้น")
}

func seedProductsAndVariants(db *sqlx.DB) {
	log.Println("...กำลังดึงข้อมูล Products")

	// ดึงสินค้ามา 30 ชิ้น
	resp, err := http.Get("https://api.escuelajs.co/api/v1/products?limit=30&offset=0")
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

	colors := []string{"Red", "Blue", "Black", "White", "Green", "Navy", "Grey"}
	sizes := []string{"S", "M", "L", "XL", "XXL"}

	successCount := 0
	for _, p := range platziProducts {
		var localCategoryID uint
		err := db.Get(&localCategoryID, "SELECT id FROM categories WHERE name=$1 LIMIT 1", p.Category.Name)
		if err != nil {
			log.Printf("⚠️ ข้ามสินค้า '%s' (ไม่พบหมวดหมู่ %s)\n", p.Title, p.Category.Name)
			continue
		}

		var newProductID uint
		randomStock := rand.Intn(100) + 10

		// 💡 แปลง []string ให้กลายเป็น JSON เพื่อบันทึกลงคอลัมน์ images (JSONB)
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

		numVariants := rand.Intn(3) + 3
		for i := 0; i < numVariants; i++ {
			color := colors[rand.Intn(len(colors))]
			size := sizes[rand.Intn(len(sizes))]
			variantStock := rand.Intn(20) + 1
			variantPrice := p.Price

			// 💡 1. สร้างรหัส SKU ให้ไม่ซ้ำกัน (เช่น PRD-1-RED-M-0)
			colorCode := strings.ToUpper(color)
			sku := fmt.Sprintf("PRD-%d-%s-%s-%d", newProductID, colorCode, size, i)

			// 💡 2. จัดรูปแบบ Attributes ลง Map แล้วแปลงเป็น JSONB
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
		log.Printf("✔️ สร้างสินค้าสำเร็จ: %s (พร้อม Variants)\n", p.Title)
	}

	log.Printf("🎉 สร้างสินค้าพร้อมขายทั้งหมด %d รายการ!\n", successCount)
}
