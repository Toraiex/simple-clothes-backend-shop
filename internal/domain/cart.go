package domain

import "time"

// ----------------------------------------------------
// 1. Struct (โครงสร้างข้อมูล)
// ----------------------------------------------------

// Cart คือ ตะกร้าของ User 1 คน
type Cart struct {
	ID        uint      `db:"id" json:"id"`
	UserID    uint      `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	Items []CartItem `db:"-" json:"items"` // 👈 เอาไว้โชว์ของในตะกร้า
}

// CartItem คือ ของแต่ละชิ้นที่อยู่ในตะกร้า
type CartItem struct {
	ID        uint      `db:"id" json:"id"`
	CartID    uint      `db:"cart_id" json:"cart_id"`
	VariantID uint      `db:"variant_id" json:"variant_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// 💡 สองตัวนี้เอาไว้ใช้ตอน "แสดงผลหน้าเว็บ" (JOIN ตารางมาโชว์ลูกค้า)
	Variant *ProductVariant `db:"-" json:"variant,omitempty"`
	Product *Product        `db:"-" json:"product,omitempty"`
}

// ----------------------------------------------------
// 2. Interfaces (สัญญาจ้าง)
// ----------------------------------------------------

type CartRepository interface {
	// หาตะกร้าของ User คนนี้ (ถ้าไม่มีต้องสร้างใหม่)
	GetCartByUserID(userID uint) (*Cart, error)
	CreateCart(userID uint) (*Cart, error)

	// จัดการของในตะกร้า
	AddItem(cartID uint, variantID uint, quantity int) error
	UpdateItemQuantity(cartItemID uint, quantity int) error
	RemoveItem(cartItemID uint) error
	ClearCart(cartID uint) error // ล้างตะกร้า (มักใช้ตอนกดสั่งซื้อเสร็จแล้ว)

	// ดึงรายการของในตะกร้าพร้อมรายละเอียดสินค้า
	GetCartItemsWithDetails(cartID uint) ([]CartItem, error)
}

type CartService interface {
	// หน้าที่ของ Service ที่ให้ Controller เรียกใช้
	GetMyCart(userID uint) (*Cart, error)
	AddToCart(userID uint, variantID uint, quantity int) error
	UpdateQuantity(userID uint, cartItemID uint, quantity int) error
	RemoveFromCart(userID uint, cartItemID uint) error
}
