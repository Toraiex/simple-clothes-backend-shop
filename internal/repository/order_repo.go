package repository

import (
	"fmt"
	"simple-clothes-shop/internal/domain"
	"time"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

type OrderModel struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	Status    string
	Total     float64
	Items     []OrderItemModel `gorm:"foreignKey:OrderID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (OrderModel) TableName() string {
	return "orders"
}

type OrderItemModel struct {
	ID        uint `gorm:"primaryKey"`
	OrderID   uint
	ProductID uint
	Quantity  int
	UnitPrice float64
	Total     float64
}

func (OrderItemModel) TableName() string {
	return "order_items"
}
func (r *orderRepository) Create(order *domain.Order) error {

	return r.db.Transaction(func(tx *gorm.DB) error {

		model := OrderModel{
			UserID: order.UserID,
			Status: string(order.Status),
			Total:  order.Total,
		}

		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		for _, item := range order.Items {
			itemModel := OrderItemModel{
				OrderID:   model.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Total:     item.Total,
			}

			if err := tx.Create(&itemModel).Error; err != nil {
				fmt.Println("ITEM INSERT ERROR:", err)
				return err
			}
		}

		return nil
	})
}

func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {
	var model OrderModel

	if err := r.db.Preload("Items").First(&model, id).Error; err != nil {
		return nil, err
	}

	var items []domain.OrderItem
	for _, i := range model.Items {
		items = append(items, domain.OrderItem{
			ID:        i.ID,
			OrderID:   i.OrderID,
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			UnitPrice: i.UnitPrice,
			Total:     i.Total,
		})
	}

	order := &domain.Order{
		ID:     model.ID,
		UserID: model.UserID,
		Status: domain.OrderStatus(model.Status),
		Items:  items,
		Total:  model.Total,
	}

	return order, nil
}
func (r *orderRepository) GetByUserID(userID uint) ([]domain.Order, error) {
	var models []OrderModel

	if err := r.db.Where("user_id = ?", userID).
		Preload("Items").
		Find(&models).Error; err != nil {
		return nil, err
	}

	var orders []domain.Order

	for _, m := range models {
		var items []domain.OrderItem
		for _, i := range m.Items {
			items = append(items, domain.OrderItem{
				ID:        i.ID,
				OrderID:   i.OrderID,
				ProductID: i.ProductID,
				Quantity:  i.Quantity,
				UnitPrice: i.UnitPrice,
				Total:     i.Total,
			})
		}

		orders = append(orders, domain.Order{
			ID:     m.ID,
			UserID: m.UserID,
			Status: domain.OrderStatus(m.Status),
			Items:  items,
			Total:  m.Total,
		})
	}

	return orders, nil
}
func (r *orderRepository) UpdateStatus(id uint, status domain.OrderStatus) error {
	return r.db.Model(&OrderModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}
