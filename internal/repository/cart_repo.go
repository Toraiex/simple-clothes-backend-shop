package repository

import (
	"database/sql"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type cartRepository struct {
	db *sqlx.DB
}

// เช็คว่า cartRepository implement CartRepository interface ครบถ้วนหรือไม่
var _ domain.CartRepository = (*cartRepository)(nil)

func NewCartRepository(db *sqlx.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

// -----------------------------------------------------------------
// 1. จัดการตัวตะกร้าหลัก (Cart)
// -----------------------------------------------------------------

func (r *cartRepository) GetCartByUserID(userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.Get(&cart, `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1`, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // คืนค่า nil ถ้ายังไม่มีตะกร้า (เดี๋ยว Service จะสั่งสร้างให้เอง)
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	// สร้างเสร็จแล้ว RETURNING เอาค่าทั้งหมดกลับมาใส่ struct ทันที
	err := r.db.Get(&cart, `
		INSERT INTO carts (user_id) VALUES ($1) 
		RETURNING id, user_id, created_at, updated_at
	`, userID)
	return &cart, err
}

// -----------------------------------------------------------------
// 2. จัดการของในตะกร้า (Cart Items)
// -----------------------------------------------------------------

func (r *cartRepository) AddItem(cartID uint, variantID uint, quantity int) error {
	// 💡 ไฮไลท์ความเจ๋ง: เราใช้ ON CONFLICT เพื่อทำ "UPSERT"
	// ถ้ายังไม่มีสินค้านี้ในตะกร้า -> INSERT
	// ถ้ามีสินค้านี้ในตะกร้าอยู่แล้ว -> UPDATE เอา quantity เดิมมาบวกของใหม่
	_, err := r.db.Exec(`
		INSERT INTO cart_items (cart_id, variant_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, variant_id) 
		DO UPDATE SET 
			quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`, cartID, variantID, quantity)

	return err
}

func (r *cartRepository) UpdateItemQuantity(cartItemID uint, quantity int) error {
	_, err := r.db.Exec(`
		UPDATE cart_items 
		SET quantity = $1, updated_at = NOW() 
		WHERE id = $2
	`, quantity, cartItemID)
	return err
}

func (r *cartRepository) RemoveItem(cartItemID uint) error {
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE id = $1`, cartItemID)
	return err
}

func (r *cartRepository) ClearCart(cartID uint) error {
	_, err := r.db.Exec(`DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}

// -----------------------------------------------------------------
// 3. ดึงข้อมูลตะกร้าแบบจัดเต็ม (พร้อมรูปและราคา) เอาไว้โชว์หน้าเว็บ
// -----------------------------------------------------------------

func (r *cartRepository) GetCartItemsWithDetails(cartID uint) ([]domain.CartItem, error) {
	// 💡 ใช้ JOIN ดึงของจาก 3 ตารางรวดเดียว: cart_items + product_variants + products
	rows, err := r.db.Query(`
		SELECT 
			ci.id, ci.cart_id, ci.variant_id, ci.quantity, ci.created_at, ci.updated_at,
			v.id, v.product_id, v.sku, v.price, v.stock, v.attributes,
			p.id, p.name, p.images
		FROM cart_items ci
		JOIN product_variants v ON ci.variant_id = v.id
		JOIN products p ON v.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC
	`, cartID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CartItem

	for rows.Next() {
		var item domain.CartItem
		var variant domain.ProductVariant
		var product domain.Product

		// แกะกล่องข้อมูลยัดใส่ทีละตัวแปรตามลำดับ SELECT ด้านบน
		err := rows.Scan(
			&item.ID, &item.CartID, &item.VariantID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Price, &variant.Stock, &variant.Attributes,
			&product.ID, &product.Name, &product.Images,
		)
		if err != nil {
			return nil, err
		}

		// เอา Variant และ Product เสียบเข้าไปใน Item
		item.Variant = &variant
		item.Product = &product

		items = append(items, item)
	}

	return items, nil
}
