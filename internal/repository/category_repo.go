package repository

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type categoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category

	err := r.db.SelectContext(ctx, &categories, `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY id DESC
	`)

	return categories, err
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category

	err := r.db.GetContext(ctx, &category, `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`,
		category.Name,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET name=$1, updated_at=NOW()
		WHERE id=$2
	`,
		category.Name,
		category.ID,
	)

	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM categories
		WHERE id=$1
	`, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("ไม่พบข้อมูลหมวดหมู่นี้ในระบบ (ลบไม่สำเร็จ)")
	}

	return nil
}
