package domain

import "time"

type Product struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	CategoryID  uint
	Image       string
	Variants    []ProductVariant
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductVariant struct {
	ID        uint
	ProductID uint
	Color     string
	Size      string
	Price     float64
	Stock     int
	CreatedAt time.Time
	UpdatedAt time.Time
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
