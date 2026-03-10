package postgres

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

type productRepository struct {
	db *sqlx.DB
}

var _ domain.ProductRepository = (*productRepository)(nil)

func NewProductRepository(db *sqlx.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE stock > 0
		ORDER BY id DESC
	`)
	if err != nil {
		logrus.Error(err) // 💡 2. ดัก Log
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE id=$1
	`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound // 💡 3. แปลงเป็นภาษากลาง
		}
		logrus.Error(err)
		return nil, err
	}

	var variants []domain.ProductVariant
	err = r.db.SelectContext(ctx, &variants, `
		SELECT id, product_id, sku, attributes, price, stock, created_at, updated_at
		FROM product_variants
		WHERE product_id=$1
	`, id)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	product.Variants = variants
	return &product, nil
}

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE category_id=$1 AND stock > 0
		ORDER BY id DESC
	`, categoryID)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return products, nil
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO products
		(name,
		description,
		price, 
		stock,
		category_id,
		images)
		VALUES ($1,
		$2,
		$3,
		$4,
		$5,
		$6) 
		RETURNING id, 
		created_at, 
		updated_at
	`,
		product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.Images,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		tx.Rollback()
		logrus.Error(err) // 💡 ดัก Log ก่อน Rollback
		return err
	}

	for i := range product.Variants {
		err := tx.QueryRowContext(ctx, `
			INSERT INTO product_variants
			(product_id, 
			sku, 
			attributes, 
			price, 
			stock)
			VALUES ($1, 
			$2,
			$3, 
			$4, 
			$5)
			RETURNING id, created_at, updated_at
		`,
			product.ID, product.Variants[i].SKU, product.Variants[i].Attributes,
			product.Variants[i].Price, product.Variants[i].Stock,
		).Scan(&product.Variants[i].ID, &product.Variants[i].CreatedAt, &product.Variants[i].UpdatedAt)

		if err != nil {
			if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {

				return fmt.Errorf("รหัส SKU นี้มีอยู่ในระบบแล้ว กรุณาใช้รหัสอื่น: %w", domain.ErrConflict)
			}
			tx.Rollback()
			logrus.Error(err)
			return errors.New("ไม่สามารถบันทึกข้อมูลได้: รหัส SKU อาจซ้ำ หรือข้อมูลไม่ถูกต้อง")
		}
	}

	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *productRepository) Update(ctx context.Context, id uint, product *domain.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = tx.QueryRowContext(ctx, `
        UPDATE products
        SET name=$1, description=$2, price=$3, stock=$4, category_id=$5, images=$6, updated_at=NOW()
        WHERE id=$7
        RETURNING updated_at
    `,
		product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.Images, id,
	).Scan(&product.UpdatedAt)

	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		logrus.Error(err)
		return err
	}

	for i := range product.Variants {
		v := &product.Variants[i]

		if v.ID == 0 {
			err := tx.QueryRowContext(ctx, `
                INSERT INTO product_variants (product_id, sku, attributes, price, stock)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id, created_at, updated_at
            `, id, v.SKU, v.Attributes, v.Price, v.Stock).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)

			if err != nil {
				tx.Rollback()
				logrus.Error(err)
				return err
			}
		} else {
			err := tx.QueryRowContext(ctx, `
                UPDATE product_variants
                SET sku=$1, attributes=$2, price=$3, stock=$4, updated_at=NOW()
                WHERE id=$5 AND product_id=$6
                RETURNING updated_at
            `, v.SKU, v.Attributes, v.Price, v.Stock, v.ID, id).Scan(&v.UpdatedAt)

			if err != nil {
				tx.Rollback()
				logrus.Error(err)
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// 💡 4. เช็คว่ามีแถวถูกลบจริงๆ ไหม (ถ้าสั่ง Delete ID มั่วๆ จะได้พ่น Not Found)
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *productRepository) DeleteVariant(ctx context.Context, variantID uint) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM product_variants WHERE id=$1`, variantID)
	if err != nil {
		logrus.Error(err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *productRepository) GetWithFilter(
	ctx context.Context,
	categoryID *uint,
	minPrice *float64,
	maxPrice *float64,
	limit int,
	offset int,
) ([]domain.Product, error) {

	query := `
		SELECT id, name, description, price, stock,
		    category_id,
			images,
			created_at,
			updated_at
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
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		logrus.Error(err) // 💡 ดัก Log
		return nil, err
	}
	return products, nil
}
