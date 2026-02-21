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
		       category_id, image, created_at, updated_at
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
		       category_id, image, created_at, updated_at
		FROM products
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	var variants []domain.ProductVariant

	err = r.db.Select(&variants, `
		SELECT id, product_id, color, size, price, stock
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
		       category_id, image, created_at, updated_at
		FROM products
		WHERE category_id=$1 AND stock > 0
		ORDER BY id DESC
	`, categoryID)

	return products, err
}

func (r *productRepository) Create(product *domain.Product) error {

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	err = tx.QueryRow(`
		INSERT INTO products
		(name, description, price, stock, category_id, image)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		product.Image,
	).Scan(&product.ID)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, v := range product.Variants {
		_, err := tx.Exec(`
			INSERT INTO product_variants
			(product_id, color, size, price, stock)
			VALUES ($1, $2, $3, $4, $5)
		`,
			product.ID,
			v.Color,
			v.Size,
			v.Price,
			v.Stock,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *productRepository) Update(id uint, product *domain.Product) error {
	// 1. เริ่ม Transaction (เพราะเราต้องแก้หลายตาราง)
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	// 2. อัปเดตข้อมูลสินค้าหลัก (Product)
	// ใช้ COALESCE หรือเช็คก่อนอัปเดต ถ้าอยากทำ Patch แบบละเอียด แต่ในที่นี้ Update หมดตามฟิลด์ที่ส่งมา
	_, err = tx.Exec(`
        UPDATE products
        SET name=$1, description=$2, price=$3, stock=$4, category_id=$5, image=$6, updated_at=NOW()
        WHERE id=$7
    `,
		product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.Image, id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. จัดการ Variants (Loop เช็คทีละตัว)
	for _, v := range product.Variants {
		if v.ID == 0 {
			// ✅ กรณีที่ 1: ไม่มี ID ส่งมา = "สร้าง Variant ใหม่" (Insert)
			_, err := tx.Exec(`
                INSERT INTO product_variants (product_id, color, size, price, stock)
                VALUES ($1, $2, $3, $4, $5)
            `, id, v.Color, v.Size, v.Price, v.Stock)

			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// ✅ กรณีที่ 2: มี ID ส่งมา = "แก้ไข Variant เดิม" (Update)
			_, err := tx.Exec(`
                UPDATE product_variants
                SET color=$1, size=$2, price=$3, stock=$4, updated_at=NOW()
                WHERE id=$5 AND product_id=$6
            `, v.Color, v.Size, v.Price, v.Stock, v.ID, id)

			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 4. จบงาน
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
) ([]domain.Product, error) {

	query := `
		SELECT id, name, description, price, stock,
		       category_id, image, created_at, updated_at
		FROM products
		WHERE stock > 0
	`

	args := []interface{}{}
	argID := 1

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

	query += " ORDER BY id DESC"

	var products []domain.Product

	err := r.db.Select(&products, query, args...)
	return products, err
}
