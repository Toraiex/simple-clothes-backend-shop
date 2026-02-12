package repository

import (
	"simple-clothes-shop/internal/domain" // เรียกใช้กฎจาก Domain
	"time"

	"gorm.io/gorm"
)

// สร้าง Struct เก็บ Database Connection
type productRepository struct {
	db *gorm.DB
}

type ProductModel struct {
	ID          uint `gorm:"primaryKey"`
	Name        string
	Description string
	Price       float64
	Stock       int
	CategoryID  uint
	Image       string
	Variants    []ProductVariantModel `gorm:"foreignKey:ProductID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductVariantModel struct {
	ID        uint `gorm:"primaryKey"`
	ProductID uint
	Color     string
	Size      string
	Price     float64
	Stock     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ProductModel) TableName() string {
	return "products"
}
func (ProductVariantModel) TableName() string {
	return "product_variants"
}

// ฟังก์ชันสร้างคนงานใหม่ (NewProductRepository)
// รับ DB เข้ามา -> ส่งคืน Interface ออกไป (Dependency Injection)
func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

// ==========================================
// เริ่มทำงานตามสั่ง (Implement Interface)
// ==========================================

// 1. ดึงสินค้าทั้งหมด
func (r *productRepository) GetAll() ([]domain.Product, error) {
	var products []domain.Product
	// ใช้ Preload("Variants") เพื่อดึงสี/ไซส์ มาด้วยเสมอ
	err := r.db.Preload("Variants").Find(&products).Error
	return products, err
}

// 2. ดึงสินค้าตาม ID
func (r *productRepository) GetByID(id uint) (*domain.Product, error) {
	var model ProductModel

	if err := r.db.
		Preload("Variants").
		First(&model, id).Error; err != nil {
		return nil, err
	}

	product := toDomainProduct(model)
	return &product, nil
}

// 3. สร้างสินค้าใหม่
func (r *productRepository) Create(product *domain.Product) error {
	// GORM ฉลาดพอที่จะบันทึก Variants (ลูก) ให้ด้วย ถ้าเราส่งมาใน Struct
	return r.db.Create(product).Error
}

// 4. อัปเดตสินค้า
func (r *productRepository) Update(id uint, product *domain.Product) error {
	product.ID = id
	// ใช้ Save เพื่ออัปเดตทุกฟิลด์ หรือ Updates สำหรับเฉพาะฟิลด์ที่ส่งมา
	return r.db.Model(&domain.Product{}).
		Where("id = ?", id).
		Updates(product).Error

}

// 5. ลบสินค้า
func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}
