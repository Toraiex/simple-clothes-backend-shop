package repository

import (
	"simple-clothes-shop/internal/domain"

	"gorm.io/gorm"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category
	// ⚡ Preload("Products") จะสั่งให้ DB ไปดึงสินค้าที่อยู่ในหมวดนั้นๆ มาใส่ใน Slice ให้เลย
	err := r.db.Preload("Products").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByID(id uint) (*domain.Category, error) {
	var category domain.Category
	// ⚡ ดึงหมวดหมู่เดียว พร้อมสินค้าทั้งหมดในหมวดนั้น
	err := r.db.Preload("Products").First(&category, id).Error
	return &category, err
}

// เพิ่มใน Interface
// GetCategory(id uint) (*domain.Category, error)

func (r *categoryRepository) Create(category *domain.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) Update(id uint, category *domain.Category) error {
	category.ID = id
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Category{}, id).Error
}
