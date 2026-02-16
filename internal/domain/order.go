package domain

import "time"

type OrderStatus string

const (
	StatusPending  OrderStatus = "pending"
	StatusPaid     OrderStatus = "paid"
	StatusShipped  OrderStatus = "shipped"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID        uint        `db:"id"`
	UserID    uint        `db:"user_id"`
	Status    OrderStatus `db:"status"`
	Total     float64     `db:"total"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`

	Items []OrderItem `db:"-"`
}

type OrderItem struct {
	ID        uint    `db:"id"`
	OrderID   uint    `db:"order_id"`
	ProductID uint    `db:"product_id"`
	Quantity  int     `db:"quantity"`
	UnitPrice float64 `db:"unit_price"`
	Total     float64 `db:"total"`
}

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id uint) (*Order, error)
	GetByUserID(userID uint) ([]Order, error)
	UpdateStatus(id uint, status OrderStatus) error
}
type OrderService interface {
	CreateOrder(userID uint, items []OrderItem) error
	GetByUserID(userID uint) ([]Order, error)
	GetByID(id uint) (*Order, error)
	GetByIDForUser(userID uint, orderID uint) (*Order, error)
	CancelOrder(userID uint, orderID uint) error
	AdminUpdateStatus(orderID uint, status OrderStatus) error
}
