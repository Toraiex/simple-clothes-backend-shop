package domain

import (
	"context" // 👈 เพิ่ม context
	"time"
)

type Cart struct {
	ID        uint      `db:"id" json:"id"`
	UserID    uint      `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Items []CartItem `db:"-" json:"items"`
}

type CartItem struct {
	ID        uint      `db:"id" json:"id"`
	CartID    uint      `db:"cart_id" json:"cart_id"`
	VariantID uint      `db:"variant_id" json:"variant_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Variant *ProductVariant `db:"-" json:"variant,omitempty"`
	Product *Product        `db:"-" json:"product,omitempty"`
}

type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID uint) (*Cart, error)
	CreateCart(ctx context.Context, userID uint) (*Cart, error)

	AddItem(ctx context.Context, cartID uint, variantID uint, quantity int) error
	UpdateItemQuantity(ctx context.Context, cartItemID uint, quantity int) error
	RemoveItem(ctx context.Context, cartItemID uint) error
	ClearCart(ctx context.Context, cartID uint) error

	GetCartItemsWithDetails(ctx context.Context, cartID uint) ([]CartItem, error)
}

type CartService interface {
	GetMyCart(ctx context.Context, userID uint) (*Cart, error)
	AddToCart(ctx context.Context, userID uint, variantID uint, quantity int) error
	UpdateQuantity(ctx context.Context, userID uint, cartItemID uint, quantity int) error
	RemoveFromCart(ctx context.Context, userID uint, cartItemID uint) error
}
