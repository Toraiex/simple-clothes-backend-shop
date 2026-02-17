package repository

import (
	"errors"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}
func (r *orderRepository) Create(order *domain.Order) error {

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	// 🔥 STEP 1: ลด stock ก่อนสร้าง order
	for _, item := range order.Items {

		res, err := tx.Exec(`
			UPDATE products
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
		`,
			item.Quantity,
			item.ProductID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		if rows == 0 {
			tx.Rollback()
			return errors.New("สินค้าในสต็อกไม่เพียงพอ")
		}
	}

	// 🔥 STEP 2: สร้าง order
	var orderID uint

	err = tx.QueryRow(`
		INSERT INTO orders (user_id, status, total)
		VALUES ($1, $2, $3)
		RETURNING id
	`,
		order.UserID,
		string(order.Status),
		order.Total,
	).Scan(&orderID)

	if err != nil {
		tx.Rollback()
		return err
	}

	// 🔥 STEP 3: สร้าง order_items
	for _, item := range order.Items {
		_, err := tx.Exec(`
			INSERT INTO order_items
			(order_id, product_id, quantity, unit_price, total)
			VALUES ($1, $2, $3, $4, $5)
		`,
			orderID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.Total,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {

	var order domain.Order

	err := r.db.Get(&order, `
		SELECT id, user_id, status, total
		FROM orders
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	var items []domain.OrderItem

	err = r.db.Select(&items, `
		SELECT id, order_id, product_id, quantity, unit_price, total
		FROM order_items
		WHERE order_id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	order.Items = items

	return &order, nil
}
func (r *orderRepository) GetByUserID(userID uint) ([]domain.Order, error) {

	var orders []domain.Order

	err := r.db.Select(&orders, `
		SELECT id, user_id, status, total
		FROM orders
		WHERE user_id=$1
		ORDER BY id DESC
	`, userID)

	if err != nil {
		return nil, err
	}

	for i := range orders {

		var items []domain.OrderItem

		err := r.db.Select(&items, `
			SELECT id, order_id, product_id, quantity, unit_price, total
			FROM order_items
			WHERE order_id=$1
		`, orders[i].ID)

		if err != nil {
			return nil, err
		}

		orders[i].Items = items
	}

	return orders, nil
}
func (r *orderRepository) UpdateStatus(id uint, status domain.OrderStatus) error {

	_, err := r.db.Exec(`
		UPDATE orders
		SET status=$1
		WHERE id=$2
	`,
		string(status),
		id,
	)

	return err
}
func (r *orderRepository) CancelAndRestoreStock(orderID uint) error {

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	// 1️⃣ โหลด items ของ order
	var items []domain.OrderItem
	err = tx.Select(&items, `
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id=$1
	`, orderID)

	if err != nil {
		tx.Rollback()
		return err
	}

	// 2️⃣ คืน stock
	for _, item := range items {

		_, err := tx.Exec(`
			UPDATE products
			SET stock = stock + $1
			WHERE id = $2
		`,
			item.Quantity,
			item.ProductID,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3️⃣ update status
	_, err = tx.Exec(`
		UPDATE orders
		SET status='canceled'
		WHERE id=$1
	`, orderID)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
