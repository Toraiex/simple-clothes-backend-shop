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
	ID        uint
	UserID    uint
	Status    OrderStatus
	Items     []OrderItem
	Total     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	ID        uint
	OrderID   uint
	ProductID uint
	Quantity  int
	UnitPrice float64
	Total     float64
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
