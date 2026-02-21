package repository

import (
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

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category

	err := r.db.Select(&categories, `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY id DESC
	`)

	return categories, err
}

func (r *categoryRepository) GetByID(id uint) (*domain.Category, error) {
	var category domain.Category

	err := r.db.Get(&category, `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Create(category *domain.Category) error {
	return r.db.QueryRow(`
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`,
		category.Name,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *categoryRepository) Update(category *domain.Category) error {
	_, err := r.db.Exec(`
		UPDATE categories
		SET name=$1, updated_at=NOW()
		WHERE id=$2
	`,
		category.Name,
		category.ID,
	)

	return err
}

func (r *categoryRepository) Delete(id uint) error {
	// 1. สั่งรันคำสั่งลบ และเก็บผลลัพธ์ (result) ไว้
	result, err := r.db.Exec(`
		DELETE FROM categories
		WHERE id=$1
	`, id)

	if err != nil {
		return err // Error จาก Database (เช่น เน็ตหลุด, syntax ผิด)
	}

	// 2. ถาม Database ว่า "ตกลงเมื่อกี้ลบไปกี่แถว?"
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// 3. ถ้าลบไป 0 แถว แปลว่าไม่เจอ ID นี้ในระบบแต่แรก
	if rowsAffected == 0 {
		return errors.New("ไม่พบข้อมูลหมวดหมู่นี้ในระบบ (ลบไม่สำเร็จ)")
	}

	return nil
}
