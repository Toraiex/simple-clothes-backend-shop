package repository

import (
	"context" // 👈 เพิ่ม context
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

func (r *productRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE stock > 0
		ORDER BY id DESC
	`)
	return products, err
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
		return nil, err
	}

	var variants []domain.ProductVariant
	err = r.db.SelectContext(ctx, &variants, `
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

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
		FROM products
		WHERE category_id=$1 AND stock > 0
		ORDER BY id DESC
	`, categoryID)
	return products, err
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	// 🚀 ใช้ BeginTxx เพื่อให้ Transaction รับ Context ได้
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO products
		(name, description, price, stock, category_id, images)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`,
		product.Name, product.Description, product.Price,
		product.Stock, product.CategoryID, product.Images,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		tx.Rollback()
		return err
	}

	for i := range product.Variants {
		err := tx.QueryRowContext(ctx, `
			INSERT INTO product_variants
			(product_id, sku, attributes, price, stock)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`,
			product.ID, product.Variants[i].SKU, product.Variants[i].Attributes,
			product.Variants[i].Price, product.Variants[i].Stock,
		).Scan(&product.Variants[i].ID, &product.Variants[i].CreatedAt, &product.Variants[i].UpdatedAt)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *productRepository) Update(ctx context.Context, id uint, product *domain.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
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
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

func (r *productRepository) DeleteVariant(ctx context.Context, variantID uint) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product_variants WHERE id=$1`, variantID)
	return err
}

func (r *productRepository) GetWithFilter(
	ctx context.Context, // 👈 เพิ่ม ctx
	categoryID *uint,
	minPrice *float64,
	maxPrice *float64,
	limit int,
	offset int,
) ([]domain.Product, error) {

	query := `
		SELECT id, name, description, price, stock,
		       category_id, images, created_at, updated_at
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
	return products, err
}
