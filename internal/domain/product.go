package domain

import (
	"context" // 👈 เพิ่ม context
	"time"

	"simple-clothes-shop/pkg/datatype"
)

type Product struct {
	ID          uint                     `db:"id" json:"id"`
	Name        string                   `db:"name" json:"name"`
	Description string                   `db:"description" json:"description"`
	Price       float64                  `db:"price" json:"price"`
	Stock       int                      `db:"stock" json:"stock"`
	CategoryID  uint                     `db:"category_id" json:"category_id"`
	Images      datatype.JSONStringArray `db:"images" json:"images"`
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
	Attributes datatype.JSONAttributes `db:"attributes" json:"attributes"`
	CreatedAt  time.Time               `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time               `db:"updated_at" json:"updated_at"`
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id uint) (*Product, error)
	GetByCategoryID(ctx context.Context, categoryID uint) ([]Product, error)
	GetWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, limit int, offset int) ([]Product, error)
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, id uint, product *Product) error
	Delete(ctx context.Context, id uint) error
	DeleteVariant(ctx context.Context, id uint) error
}

type ProductUsecase interface {
	FetchAll(ctx context.Context) ([]Product, error)
	FetchByID(ctx context.Context, id uint) (*Product, error)
	FetchByCategoryID(ctx context.Context, categoryID uint) ([]Product, error)
	FetchWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, page int, limit int) ([]Product, error)
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, id uint, product *Product) error
	RemoveProduct(ctx context.Context, id uint) error
	RemoveVariant(ctx context.Context, variantID uint) error
}
