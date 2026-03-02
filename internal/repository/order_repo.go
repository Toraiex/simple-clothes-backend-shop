package repository

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	db *sqlx.DB
}

var _ domain.OrderRepository = (*orderRepository)(nil)

func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrderFromCart(ctx context.Context, userID uint, cartItems []domain.CartItem, totalAmount float64) (*domain.Order, error) {
	// 🚀 ใช้ BeginTxx สำหรับ Transaction ที่รองรับ Context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, item := range cartItems {
		// 🚀 เปลี่ยนเป็น ExecContext
		res, err := tx.ExecContext(ctx, `
			UPDATE product_variants
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
		`, item.Quantity, item.VariantID)

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		if rows == 0 {
			tx.Rollback()
			errMsg := fmt.Sprintf("สินค้า '%s' สี %s ไซส์ %s ในสต็อกมีไม่เพียงพอ",
				item.Product.Name, item.Variant.Attributes["Color"], item.Variant.Attributes["Size"])
			return nil, errors.New(errMsg)
		}
	}

	var orderID uint
	// 🚀 เปลี่ยนเป็น QueryRowContext
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, status, total)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, string(domain.StatusPending), totalAmount).Scan(&orderID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range cartItems {
		unitPrice := item.Variant.Price
		itemTotal := unitPrice * float64(item.Quantity)

		// 🚀 เปลี่ยนเป็น ExecContext
		_, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, variant_id, quantity, unit_price, total)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, item.VariantID, item.Quantity, unitPrice, itemTotal)

		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 🚀 เปลี่ยนเป็น ExecContext
	_, err = tx.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)
	`, userID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &domain.Order{
		ID:     orderID,
		UserID: userID,
		Status: domain.StatusPending,
		Total:  totalAmount,
	}, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	var order domain.Order
	// 🚀 เปลี่ยนเป็น GetContext
	err := r.db.GetContext(ctx, &order, `SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}

	var items []domain.OrderItem
	// 🚀 เปลี่ยนเป็น SelectContext
	err = r.db.SelectContext(ctx, &items, `SELECT id, order_id, variant_id, quantity, unit_price, total FROM order_items WHERE order_id=$1`, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.SelectContext(ctx, &orders, `SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE user_id=$1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		var items []domain.OrderItem
		err := r.db.SelectContext(ctx, &items, `SELECT id, order_id, variant_id, quantity, unit_price, total FROM order_items WHERE order_id=$1`, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}
	return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status domain.OrderStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2`, string(status), id)
	return err
}

func (r *orderRepository) CancelAndRestoreStock(ctx context.Context, orderID uint) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	var items []domain.OrderItem
	err = tx.SelectContext(ctx, &items, `SELECT variant_id, quantity FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, item := range items {
		_, err := tx.ExecContext(ctx, `UPDATE product_variants SET stock = stock + $1 WHERE id = $2`, item.Quantity, item.VariantID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `UPDATE orders SET status='canceled', updated_at=NOW() WHERE id=$1`, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
