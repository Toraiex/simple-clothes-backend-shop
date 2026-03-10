package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus" // 💡 1. Import logrus
)

type cartRepository struct {
	db *sqlx.DB
}

var _ domain.CartRepository = (*cartRepository)(nil)

func NewCartRepository(db *sqlx.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 💡 กรณีนี้ไม่ถือว่าพัง เพราะลูกค้าใหม่อาจจะยังไม่มีตะกร้า Usecase จะสร้างให้เอง
		}
		logrus.Error(err) // ดัก Log ถ้า DB มีปัญหา
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) CreateCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.GetContext(ctx, &cart, `
		INSERT INTO carts (user_id) VALUES ($1) 
		RETURNING id, user_id, created_at, updated_at
	`, userID)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID uint, variantID uint, quantity int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cart_items (cart_id, 
		variant_id, 
		quantity)
		VALUES ($1, 
		$2, 
		$3)
		ON CONFLICT (cart_id, 
		variant_id) 
		DO UPDATE SET 
			quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`, cartID, variantID, quantity)

	if err != nil {

		if strings.Contains(err.Error(), "23503") || strings.Contains(err.Error(), "foreign key constraint") {

			return fmt.Errorf("ไม่พบสินค้ารหัสนี้ในระบบ (อาจถูกลบไปแล้ว หรือไม่มีอยู่จริง): %w", domain.ErrNotFound)
		}
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, cartItemID uint, quantity int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE cart_items 
		SET quantity = $1, updated_at = NOW() 
		WHERE id = $2
	`, quantity, cartItemID)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// 💡 เช็คว่ามีการอัปเดตจริงๆ ไหม
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartItemID uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE id = $1`, cartItemID)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// 💡 เช็คว่ามีการลบจริงๆ ไหม
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *cartRepository) GetCartItemsWithDetails(ctx context.Context, cartID uint) ([]domain.CartItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			ci.id, 
			ci.cart_id, 
			ci.variant_id, 
			ci.quantity, 
			ci.created_at, 
			ci.updated_at,
			v.id, 
			v.product_id, 
			v.sku, 
			v.price, 
			v.stock, 
			v.attributes,
			p.id, 
			p.name, 
			p.images
		FROM cart_items ci
		JOIN product_variants v ON ci.variant_id = v.id
		JOIN products p ON v.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC
	`, cartID)

	if err != nil {
		logrus.Error(err)
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
			logrus.Error(err)
			return nil, err
		}

		item.Variant = &variant
		item.Product = &product
		items = append(items, item)
	}

	return items, nil
}
