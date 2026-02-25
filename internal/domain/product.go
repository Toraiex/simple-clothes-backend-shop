package domain

import (
	"time"

	"simple-clothes-shop/pkg/datatype" // 👈 Import เครื่องมือของเราเข้ามา
)

type Product struct {
	ID          uint                     `db:"id" json:"id"`
	Name        string                   `db:"name" json:"name"`
	Description string                   `db:"description" json:"description"`
	Price       float64                  `db:"price" json:"price"`
	Stock       int                      `db:"stock" json:"stock"`
	CategoryID  uint                     `db:"category_id" json:"category_id"`
	Images      datatype.JSONStringArray `db:"images" json:"images"` // 👈 ใช้ Type จาก pkg
	CreatedAt   time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time                `db:"updated_at" json:"updated_at"`

	Category *Category        `db:"-" json:"category,omitempty"`
	Variants []ProductVariant `db:"-" json:"variants,omitempty"`
}

type ProductVariant struct {
	ID         uint                    `db:"id" json:"id"`
	ProductID  uint                    `db:"product_id" json:"product_id"`
	SKU        string                  `db:"sku" json:"sku"`
	Price      float64                 `db:"price" json:"price"`
	Stock      int                     `db:"stock" json:"stock"`
	Attributes datatype.JSONAttributes `db:"attributes" json:"attributes"` // 👈 ใช้ Type จาก pkg
	CreatedAt  time.Time               `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time               `db:"updated_at" json:"updated_at"`
}

// ==========================================
// 2. Repository Interface (สัญญาจ้างฝ่ายเก็บของ)
// ==========================================
// ใครที่จะมาทำหน้าที่คุยกับ DB ต้องมีฟังก์ชันตามนี้เป๊ะๆ ห้ามขาด ห้ามเกิน
type ProductRepository interface {
	GetAll() ([]Product, error)
	GetByID(id uint) (*Product, error)
	GetByCategoryID(categoryID uint) ([]Product, error)
	GetWithFilter(categoryID *uint, minPrice *float64, maxPrice *float64, limit int, offset int) ([]Product, error)
	Create(product *Product) error
	Update(id uint, product *Product) error
	Delete(id uint) error
	DeleteVariant(id uint) error
}

type ProductService interface {
	FetchAll() ([]Product, error)
	FetchByID(id uint) (*Product, error)
	FetchByCategoryID(categoryID uint) ([]Product, error)
	FetchWithFilter(categoryID *uint, minPrice *float64, maxPrice *float64, page int, limit int) ([]Product, error)
	CreateProduct(product *Product) error
	UpdateProduct(id uint, product *Product) error
	RemoveProduct(id uint) error
	RemoveVariant(variantID uint) error
}
