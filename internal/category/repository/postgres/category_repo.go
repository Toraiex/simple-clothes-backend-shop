package repository

import (
	"context"
	"database/sql"
	"errors"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus" // 💡 1. Import logrus
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
	if err != nil {
		logrus.Error(err) // 💡 ดัก Log พัง
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category

	err := r.db.GetContext(ctx, &category, `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id=$1
	`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound // 💡 แปลงเป็นภาษากลาง
		}
		logrus.Error(err)
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`,
		category.Name,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
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

	if err != nil {
		return err
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		logrus.Error(err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
