package domain

import (
	"context" // 👈 เพิ่ม context
	"time"
)

type OrderStatus string

const (
	StatusPending  OrderStatus = "pending"
	StatusPaid     OrderStatus = "paid"
	StatusShipped  OrderStatus = "shipped"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID        uint        `db:"id" json:"id"`
	UserID    uint        `db:"user_id" json:"user_id"`
	Status    OrderStatus `db:"status" json:"status"`
	Total     float64     `db:"total" json:"total"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`

	Items []OrderItem `db:"-" json:"items"`
}

type OrderItem struct {
	ID        uint    `db:"id" json:"id"`
	OrderID   uint    `db:"order_id" json:"order_id"`
	VariantID uint    `db:"variant_id" json:"variant_id"`
	Quantity  int     `db:"quantity" json:"quantity"`
	UnitPrice float64 `db:"unit_price" json:"unit_price"`
	Total     float64 `db:"total" json:"total"`

	Variant *ProductVariant `db:"-" json:"variant,omitempty"`
	Product *Product        `db:"-" json:"product,omitempty"`
}

type OrderRepository interface {
	CreateOrderFromCart(ctx context.Context, userID uint, cartItems []CartItem, totalAmount float64) (*Order, error)
	GetByID(ctx context.Context, id uint) (*Order, error)
	GetByUserID(ctx context.Context, userID uint) ([]Order, error)
	UpdateStatus(ctx context.Context, id uint, status OrderStatus) error
	CancelAndRestoreStock(ctx context.Context, orderID uint) error
}

type OrderService interface {
	Checkout(ctx context.Context, userID uint) error
	GetByUserID(ctx context.Context, userID uint) ([]Order, error)
	GetByID(ctx context.Context, id uint) (*Order, error)
	GetByIDForUser(ctx context.Context, userID uint, orderID uint) (*Order, error)
	CancelOrder(ctx context.Context, userID uint, orderID uint) error
	AdminUpdateStatus(ctx context.Context, orderID uint, status OrderStatus) error
}
