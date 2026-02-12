package domain

import "time"

type Category struct {
	ID        uint
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository Contract
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
