package main

import (
	"encoding/json"
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
	Images      []string       `json:"images"` // เผื่ออนาคตอยากเก็บรูป
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// ตั้งค่า Seed สำหรับการสุ่ม (เพื่อให้สุ่มไม่ซ้ำเดิมทุกครั้ง)
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

// ==========================================
// ฟังก์ชัน Seed Categories
// ==========================================
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

// ==========================================
// ฟังก์ชัน Seed Products และ Variants (รองรับ Schema ใหม่)
// ==========================================
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

	// เตรียมชุดข้อมูลสำหรับสุ่ม Variants
	colors := []string{"Red", "Blue", "Black", "White", "Green", "Navy", "Grey", "Yellow"}
	sizes := []string{"S", "M", "L", "XL", "XXL"}

	successCount := 0
	for _, p := range platziProducts {
		// 1. หา ID ของ Category ใน Database ของเรา
		var localCategoryID uint
		err := db.Get(&localCategoryID, "SELECT id FROM categories WHERE name=$1 LIMIT 1", p.Category.Name)
		if err != nil {
			log.Printf("⚠️ ข้ามสินค้า '%s' (ไม่พบหมวดหมู่ %s)\n", p.Title, p.Category.Name)
			continue
		}

		// 2. Insert ลงตาราง Products
		var newProductID uint
		randomStock := rand.Intn(100) + 10 // สุ่มสต็อกรวม (จริงๆ ควรเป็นผลรวมของ variant แต่ใส่ไว้ก่อน)

		err = db.QueryRow(`
			INSERT INTO products (name, description, price, stock, category_id, image)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`, p.Title, p.Description, p.Price, randomStock, localCategoryID, p.Images).Scan(&newProductID)

		if err != nil {
			log.Printf("❌ สร้างสินค้า '%s' ไม่สำเร็จ: %v\n", p.Title, err)
			continue
		}

		// 3. Insert ลงตาราง Product Variants (สุ่มสร้าง 3-5 แบบต่อสินค้า 1 ชิ้น)
		numVariants := rand.Intn(3) + 3
		for i := 0; i < numVariants; i++ {
			// สุ่มสีและไซส์
			color := colors[rand.Intn(len(colors))]
			size := sizes[rand.Intn(len(sizes))]
			variantStock := rand.Intn(20) + 1

			// ราคา Variant อาจจะเท่ากับสินค้าหลัก หรือบวกเพิ่มนิดหน่อย (จำลองสถานการณ์จริง)
			variantPrice := p.Price

			_, vErr := db.Exec(`
				INSERT INTO product_variants (product_id, color, size, price, stock) 
				VALUES ($1, $2, $3, $4, $5)
			`, newProductID, color, size, variantPrice, variantStock)

			if vErr != nil {
				log.Printf("   ⚠️ สร้าง Variant (%s-%s) ไม่สำเร็จ: %v\n", color, size, vErr)
			}
		}

		successCount++
		log.Printf("✔️ สร้างสินค้าสำเร็จ: %s (พร้อม Variants)\n", p.Title)
	}

	log.Printf("🎉 สร้างสินค้าพร้อมขายทั้งหมด %d รายการ!\n", successCount)
}
