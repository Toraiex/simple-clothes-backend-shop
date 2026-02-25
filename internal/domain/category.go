package domain

import "time"

type Category struct {
	ID        uint      `db:"id"         json:"id"`         // 👈 เพิ่ม json:"id"
	Name      string    `db:"name"       json:"name"`       // 👈 เพิ่ม json:"name"
	CreatedAt time.Time `db:"created_at" json:"created_at"` // 👈 เพิ่ม json:"created_at"
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"` // 👈 เพิ่ม json:"updated_at"

	Products []Product `db:"-" json:"products,omitempty"`
}

type CategoryRepository interface {
	GetAll() ([]Category, error)
	GetByID(id uint) (*Category, error)
	Create(category *Category) error
	Update(category *Category) error
	Delete(id uint) error
}

// Service Contract
type CategoryService interface {
	FetchAll() ([]Category, error)
	GetCategory(id uint) (*Category, error)
	CreateCategory(name string) error
	UpdateCategory(id uint, name string) error
	RemoveCategory(id uint) error
}
