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
	GetByCategoryID(categoryID uint) ([]Product, error)
	GetWithFilter(categoryID *uint, minPrice *float64, maxPrice *float64) ([]Product, error) // ✅ เพิ่ม
	Create(product *Product) error
	Update(id uint, product *Product) error
	Delete(id uint) error
}

type ProductService interface {
	FetchAll() ([]Product, error)
	FetchByID(id uint) (*Product, error)
	FetchByCategoryID(categoryID uint) ([]Product, error)
	FetchWithFilter(categoryID *uint, minPrice *float64, maxPrice *float64) ([]Product, error) // ✅ เพิ่ม
	CreateProduct(product *Product) error
	UpdateProduct(id uint, product *Product) error
	RemoveProduct(id uint) error
}
