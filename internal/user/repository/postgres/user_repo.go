package postgres

import (
	"context"
	"database/sql"
	"errors"
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus" // 💡 สไตล์ bxcodec ใช้ logrus ใน Repo
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, password, role, address, phone, email, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Username, user.Password, user.Role,
		user.Address, user.Phone, user.Email, user.IsVerified,
	).Scan(&user.ID)

	if err != nil {
		logrus.Error(err) // 💡 ปริ้น Log ถ้า DB พัง
		return err
	}
	return nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password, role, address, phone, is_verified, email
		FROM users WHERE username=$1
	`, username)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logrus.Error(err) // 💡 ปริ้น Log ถ้าเป็น Error แปลกๆ ที่ไม่ใช่หาไม่เจอ
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `
		SELECT id, username, password, role, address, phone, is_verified, email
		FROM users WHERE id=$1
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logrus.Error(err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, id uint, user *domain.User) error {
	query := `
        UPDATE users 
        SET address=$1, role=$2, phone=$3, updated_at=NOW() 
        WHERE id=$4
    `
	_, err := r.db.ExecContext(ctx, query, user.Address, string(user.Role), user.Phone, id)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, `
		SELECT id, username, role, address, phone, is_verified, email
		FROM users ORDER BY id ASC
	`)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `
        SELECT id, username, password, role, address, phone, email, is_verified
        FROM users WHERE email = $1
    `
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logrus.Error(err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateVerificationStatus(ctx context.Context, userID uint) error {
	query := `UPDATE users SET is_verified = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, newPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, newPassword, userID)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}
