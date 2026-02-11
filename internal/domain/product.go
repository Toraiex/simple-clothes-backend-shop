package domain

import (
	"gorm.io/gorm"
)

// ==========================================
// 1. Entities (Models) - หน้าตาของข้อมูล
// ==========================================

// สินค้าหลัก
type Product struct {
	gorm.Model
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  uint    `json:"category_id"`
	Image       string  `json:"image"`
	// Relationship: 1 Product มีหลาย Variants
	Variants []ProductVariant `json:"variants" gorm:"foreignKey:ProductID"`
}

// ตัวเลือกสินค้า (สี/ไซส์)
type ProductVariant struct {
	gorm.Model
	ProductID uint    `json:"product_id"`
	Color     string  `json:"color"`
	Size      string  `json:"size"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

// ==========================================
// 2. Repository Interface (สัญญาจ้างฝ่ายเก็บของ)
// ==========================================
// ใครที่จะมาทำหน้าที่คุยกับ DB ต้องมีฟังก์ชันตามนี้เป๊ะๆ ห้ามขาด ห้ามเกิน
type ProductRepository interface {
	GetAll() ([]Product, error)
	GetByID(id uint) (*Product, error)
	Create(product *Product) error
	Update(id uint, product *Product) error
	Delete(id uint) error
}

// ==========================================
// 3. Service Interface (สัญญาจ้างฝ่ายจัดการ/สมอง)
// ==========================================
// ใครที่จะมาเป็น Business Logic ต้องมีฟังก์ชันตามนี้
type ProductService interface {
	FetchAll() ([]Product, error)
	FetchByID(id uint) (*Product, error)
	CreateProduct(product *Product) error // อาจจะมี Logic เช็คราคา หรือตัดสต็อกในนี้
	RemoveProduct(id uint) error
	UpdateProduct(id uint, product *Product) error
}
