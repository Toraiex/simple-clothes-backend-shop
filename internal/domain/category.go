package domain

import (
	"context" // 👈 เพิ่ม context
	"time"
)

type Category struct {
	ID        uint      `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Products []Product `db:"-" json:"products,omitempty"`
}

type CategoryRepository interface {
	GetAll(ctx context.Context) ([]Category, error)
	GetByID(ctx context.Context, id uint) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uint) error
}

// Service Contract
type CategoryService interface {
	FetchAll(ctx context.Context) ([]Category, error)
	GetCategory(ctx context.Context, id uint) (*Category, error)
	CreateCategory(ctx context.Context, name string) error
	UpdateCategory(ctx context.Context, id uint, name string) error
	RemoveCategory(ctx context.Context, id uint) error
}
