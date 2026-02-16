package domain

import "time"

type Product struct {
	ID          uint      `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	CategoryID  uint      `db:"category_id"` // 🔥 ตัวนี้สำคัญ
	Image       string    `db:"image"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	Variants []ProductVariant `db:"-"`
}

type ProductVariant struct {
	ID        uint      `db:"id"`
	ProductID uint      `db:"product_id"`
	Color     string    `db:"color"`
	Size      string    `db:"size"`
	Price     float64   `db:"price"`
	Stock     int       `db:"stock"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
