package repository

import (
	"fmt"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type productRepository struct {
	db *sqlx.DB
}

var _ domain.ProductRepository = (*productRepository)(nil)

func NewProductRepository(db *sqlx.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll() ([]domain.Product, error) {

	var products []domain.Product

	err := r.db.Select(&products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE stock > 0
		ORDER BY id DESC
	`)

	return products, err
}

func (r *productRepository) GetByID(id uint) (*domain.Product, error) {

	var product domain.Product

	err := r.db.Get(&product, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	var variants []domain.ProductVariant

	err = r.db.Select(&variants, `
		SELECT id, product_id, sku, attributes, price, stock, created_at, updated_at
		FROM product_variants
		WHERE product_id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	product.Variants = variants

	return &product, nil
}

func (r *productRepository) GetByCategoryID(categoryID uint) ([]domain.Product, error) {

	var products []domain.Product

	err := r.db.Select(&products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE category_id=$1 AND stock > 0
		ORDER BY id DESC
	`, categoryID)

	return products, err
}

// ==========================================
// 1. สร้างสินค้าใหม่ (แก้บัค Timestamp)
// ==========================================
func (r *productRepository) Create(product *domain.Product) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	// 1. บันทึกข้อมูล Product หลัก
	err = tx.QueryRow(`
		INSERT INTO products
		(name, description, price, stock, category_id, images)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at -- 👈 ขอ 3 ค่านี้กลับมาจาก Database
	`,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.Images,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt) // 👈 จับยัดกลับเข้า Struct ทันที

	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. บันทึกข้อมูล Variants (สี/ไซส์)
	// 💡 ต้องใช้ for i := range เพื่อให้อ้างอิงถึงตำแหน่งตัวแปรจริงๆ (Pointer) ใน Array
	for i := range product.Variants {
		err := tx.QueryRow(`
			INSERT INTO product_variants
			(product_id, sku, attributes, price, stock)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at -- 👈 ขอค่ากลับมาเหมือนกัน
		`,
			product.ID,
			product.Variants[i].SKU,
			product.Variants[i].Attributes,
			product.Variants[i].Price,
			product.Variants[i].Stock,
		).Scan(&product.Variants[i].ID, &product.Variants[i].CreatedAt, &product.Variants[i].UpdatedAt)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *productRepository) Update(id uint, product *domain.Product) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	// 1. อัปเดตสินค้าหลัก
	err = tx.QueryRow(`
        UPDATE products
        SET name=$1, description=$2, price=$3, stock=$4, category_id=$5, images=$6, updated_at=NOW()
        WHERE id=$7
        RETURNING updated_at -- 👈 ขอเวลาที่เพิ่งอัปเดตกลับมา
    `,
		product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.Images, id,
	).Scan(&product.UpdatedAt) // 👈 ยัดใส่ Struct

	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. จัดการ Variants
	for i := range product.Variants {
		v := &product.Variants[i] // ใช้ Pointer ช่วยให้โค้ดสั้นลง

		if v.ID == 0 {
			// สร้างใหม่ (Insert)
			err := tx.QueryRow(`
                INSERT INTO product_variants (product_id, sku, attributes, price, stock)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id, created_at, updated_at
            `, id, v.SKU, v.Attributes, v.Price, v.Stock).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)

			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// แก้ไขอันเดิม (Update)
			err := tx.QueryRow(`
                UPDATE product_variants
                SET sku=$1, attributes=$2, price=$3, stock=$4, updated_at=NOW()
                WHERE id=$5 AND product_id=$6
                RETURNING updated_at
            `, v.SKU, v.Attributes, v.Price, v.Stock, v.ID, id).Scan(&v.UpdatedAt)

			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *productRepository) Delete(id uint) error {

	_, err := r.db.Exec(`
		DELETE FROM products
		WHERE id=$1
	`, id)

	return err
}

// เพิ่มใน productRepository
func (r *productRepository) DeleteVariant(variantID uint) error {
	_, err := r.db.Exec(`DELETE FROM product_variants WHERE id=$1`, variantID)
	return err
}
func (r *productRepository) GetWithFilter(
	categoryID *uint,
	minPrice *float64,
	maxPrice *float64,
	limit int, // 👈 เพิ่ม Limit (จำนวนต่อหน้า)
	offset int, // 👈 เพิ่ม Offset (จุดเริ่มต้น)
) ([]domain.Product, error) {

	query := `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE stock > 0
	`

	args := []interface{}{}
	argID := 1 // เริ่มนับตัวแปรที่ $1

	// ตรวจสอบเงื่อนไขทีละข้อ
	if categoryID != nil {
		query += " AND category_id=$" + fmt.Sprint(argID)
		args = append(args, *categoryID)
		argID++
	}

	if minPrice != nil {
		query += " AND price >= $" + fmt.Sprint(argID)
		args = append(args, *minPrice)
		argID++
	}

	if maxPrice != nil {
		query += " AND price <= $" + fmt.Sprint(argID)
		args = append(args, *maxPrice)
		argID++
	}

	// เรียงลำดับจากใหม่ไปเก่า
	query += " ORDER BY id DESC"

	// 🚀 ใส่ Pagination เข้าไปท้ายสุด
	// ตัวอย่าง: LIMIT $4 OFFSET $5
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	var products []domain.Product

	// สั่งรัน Query พร้อมตัวแปรทั้งหมด
	err := r.db.Select(&products, query, args...)
	return products, err
}
