package repository

import (
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type orderRepository struct {
	db *sqlx.DB
}

// เช็คว่า implement ครบตาม interface ไหม
var _ domain.OrderRepository = (*orderRepository)(nil)

func NewOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

// =================================================================
// 1. สร้างใบสั่งซื้อจากตะกร้าสินค้า (Transaction ขั้นเทพ)
// =================================================================
func (r *orderRepository) CreateOrderFromCart(userID uint, cartItems []domain.CartItem, totalAmount float64) (*domain.Order, error) {
	// เริ่มต้น Transaction (หยุดเวลา Database)
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}

	// 🔥 STEP 1: ลดสต็อก (ไปตัดที่ product_variants แทน products)
	for _, item := range cartItems {
		res, err := tx.Exec(`
			UPDATE product_variants
			SET stock = stock - $1
			WHERE id = $2 AND stock >= $1
		`, item.Quantity, item.VariantID)

		if err != nil {
			tx.Rollback() // ถ้า Error ให้ย้อนเวลา
			return nil, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// ถ้า update ไม่ได้ (rows == 0) แปลว่าของชิ้นนั้นสต็อกไม่พอ!
		if rows == 0 {
			tx.Rollback()
			// บอกไปเลยว่าสินค้าตัวไหนหมด (ดึงชื่อมาจาก JOIN ตอน GetCartItems)
			errMsg := fmt.Sprintf("สินค้า '%s' สี %s ไซส์ %s ในสต็อกมีไม่เพียงพอ",
				item.Product.Name, item.Variant.Attributes["Color"], item.Variant.Attributes["Size"])
			return nil, errors.New(errMsg)
		}
	}

	// 🔥 STEP 2: สร้าง Order หลัก
	var orderID uint
	err = tx.QueryRow(`
		INSERT INTO orders (user_id, status, total)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, string(domain.StatusPending), totalAmount).Scan(&orderID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 🔥 STEP 3: สร้าง Order Items (คัดลอกของจากตะกร้าลงบิล)
	for _, item := range cartItems {
		unitPrice := item.Variant.Price
		itemTotal := unitPrice * float64(item.Quantity)

		_, err := tx.Exec(`
			INSERT INTO order_items (order_id, variant_id, quantity, unit_price, total)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, item.VariantID, item.Quantity, unitPrice, itemTotal)

		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 🔥 STEP 4: ล้างตะกร้าสินค้า (ลบเฉพาะของ User คนนี้)
	_, err = tx.Exec(`
		DELETE FROM cart_items
		WHERE cart_id = (SELECT id FROM carts WHERE user_id = $1)
	`, userID)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// กดยืนยันการเปลี่ยนแปลงทั้งหมด! (Commit)
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

// =================================================================
// ฟังก์ชันอื่นๆ (แก้ให้ชี้ไปที่ variant_id แทน product_id)
// =================================================================

func (r *orderRepository) GetByID(id uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Get(&order, `SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}

	var items []domain.OrderItem
	// 👈 เปลี่ยน product_id เป็น variant_id
	err = r.db.Select(&items, `SELECT id, order_id, variant_id, quantity, unit_price, total FROM order_items WHERE order_id=$1`, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return &order, nil
}

func (r *orderRepository) GetByUserID(userID uint) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Select(&orders, `SELECT id, user_id, status, total, created_at, updated_at FROM orders WHERE user_id=$1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		var items []domain.OrderItem
		// 👈 เปลี่ยน product_id เป็น variant_id
		err := r.db.Select(&items, `SELECT id, order_id, variant_id, quantity, unit_price, total FROM order_items WHERE order_id=$1`, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}
	return orders, nil
}

func (r *orderRepository) UpdateStatus(id uint, status domain.OrderStatus) error {
	_, err := r.db.Exec(`UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2`, string(status), id)
	return err
}

func (r *orderRepository) CancelAndRestoreStock(orderID uint) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	var items []domain.OrderItem
	// 👈 เปลี่ยน product_id เป็น variant_id
	err = tx.Select(&items, `SELECT variant_id, quantity FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, item := range items {
		// 👈 คืนสต็อกไปที่ตาราง product_variants
		_, err := tx.Exec(`UPDATE product_variants SET stock = stock + $1 WHERE id = $2`, item.Quantity, item.VariantID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec(`UPDATE orders SET status='canceled', updated_at=NOW() WHERE id=$1`, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
