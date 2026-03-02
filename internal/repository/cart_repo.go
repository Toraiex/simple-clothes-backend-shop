package repository

import (
	"context" // 👈 เพิ่ม context
	"database/sql"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type cartRepository struct {
	db *sqlx.DB
}

var _ domain.CartRepository = (*cartRepository)(nil)

func NewCartRepository(db *sqlx.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

// -----------------------------------------------------------------
// 1. จัดการตัวตะกร้าหลัก (Cart)
// -----------------------------------------------------------------

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	// 🚀 ใช้ GetContext
	err := r.db.GetContext(ctx, &cart, `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1`, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	// 🚀 ใช้ GetContext
	err := r.db.GetContext(ctx, &cart, `
		INSERT INTO carts (user_id) VALUES ($1) 
		RETURNING id, user_id, created_at, updated_at
	`, userID)
	return &cart, err
}

// -----------------------------------------------------------------
// 2. จัดการของในตะกร้า (Cart Items)
// -----------------------------------------------------------------

func (r *cartRepository) AddItem(ctx context.Context, cartID uint, variantID uint, quantity int) error {
	// 🚀 ใช้ ExecContext
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (cart_id, variant_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, variant_id) 
		DO UPDATE SET 
			quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`, cartID, variantID, quantity)

	return err
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartItemID uint, quantity int) error {
	// 🚀 ใช้ ExecContext
	_, err := r.db.ExecContext(ctx, `
		UPDATE cart_items 
		SET quantity = $1, updated_at = NOW() 
		WHERE id = $2
	`, quantity, cartItemID)
	return err
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartItemID uint) error {
	// 🚀 ใช้ ExecContext
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE id = $1`, cartItemID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID uint) error {
	// 🚀 ใช้ ExecContext
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}

// -----------------------------------------------------------------
// 3. ดึงข้อมูลตะกร้าแบบจัดเต็ม (พร้อมรูปและราคา) เอาไว้โชว์หน้าเว็บ
// -----------------------------------------------------------------

func (r *cartRepository) GetCartItemsWithDetails(ctx context.Context, cartID uint) ([]domain.CartItem, error) {
	// 🚀 ใช้ QueryContext
	rows, err := r.db.QueryContext(ctx, `
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

		err := rows.Scan(
			&item.ID, &item.CartID, &item.VariantID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
			&variant.ID, &variant.ProductID, &variant.SKU, &variant.Price, &variant.Stock, &variant.Attributes,
			&product.ID, &product.Name, &product.Images,
		)
		if err != nil {
			return nil, err
		}

		item.Variant = &variant
		item.Product = &product

		items = append(items, item)
	}

	return items, nil
}
