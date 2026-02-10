package models

import (
	"gorm.io/gorm"
)

// 1. User
type User struct {
	gorm.Model
	Username string  `json:"username" gorm:"unique"`
	Password string  `json:"-"`
	Role     string  `json:"role" gorm:"default:user"`
	Address  string  `json:"address"`
	Orders   []Order `json:"orders"`
}

// 2. Category
type Category struct {
	gorm.Model
	Name     string    `json:"name"`
	Products []Product `json:"products"`
}

// 3. Product
type Product struct {
	gorm.Model
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Price       float64          `json:"price"`
	Stock       int              `json:"stock"`
	CategoryID  uint             `json:"category_id"`
	Category    Category         `json:"category" gorm:"foreignKey:CategoryID"`
	Image       string           `json:"image"`
	Variants    []ProductVariant `json:"variants"`
}

// 4. Product Variant (ลูกของ Product)
type ProductVariant struct {
	gorm.Model
	ProductID uint    `json:"product_id"`
	Color     string  `json:"color"`
	Size      string  `json:"size"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

// 5. Order
type Order struct {
	gorm.Model
	UserID     uint        `json:"user_id"`
	User       User        `json:"user"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status" gorm:"default:pending"`
	OrderItems []OrderItem `json:"order_items"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// 6. Review
type Review struct {
	gorm.Model
	UserID    uint   `json:"user_id"`
	User      User   `json:"user"`
	ProductID uint   `json:"product_id"`
	Content   string `json:"content"`
	Rating    int    `json:"rating"`
}
