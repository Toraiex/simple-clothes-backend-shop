package domain

import "gorm.io/gorm"

// Category Entity
type Category struct {
	gorm.Model
	Name     string    `json:"name" gorm:"unique;not null"` // ชื่อหมวดหมู่ต้องไม่ซ้ำและไม่ว่าง
	Products []Product `json:"products,omitempty"`          // เชื่อมโยงกลับไปหา Product
}

// CategoryRepository Interface (กฎสำหรับคนงาน DB)
type CategoryRepository interface {
	GetAll() ([]Category, error)
	GetByID(id uint) (*Category, error)
	Create(category *Category) error
	Update(id uint, category *Category) error
	Delete(id uint) error
}

// CategoryService Interface (กฎสำหรับสมองของระบบ)
type CategoryService interface {
	FetchAll() ([]Category, error)
	CreateCategory(name string) error
	RemoveCategory(id uint) error
	GetCategory(id uint) (*Category, error)
	UpdateCategory(id uint, name string) error
}
