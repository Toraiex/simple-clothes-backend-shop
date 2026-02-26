package repository

import (
	"simple-clothes-shop/internal/domain"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// ==========================================
// สร้างผู้ใช้งานใหม่ (Register)
// ==========================================
func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (username, password, role, address, phone, email, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRow(query,
		user.Username, user.Password, user.Role,
		user.Address, user.Phone, user.Email, user.IsVerified,
	).Scan(&user.ID)
	return err
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {

	var user domain.User

	err := r.db.Get(&user, `
		SELECT id, username, password, role, address, phone, is_verified, email
		FROM users
		WHERE username=$1
	`, username)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(id uint) (*domain.User, error) {

	var user domain.User

	err := r.db.Get(&user, `
		SELECT id, username, password, role, address, phone, is_verified, email
		FROM users
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(id uint, user *domain.User) error {
	query := `
        UPDATE users 
        SET address=$1, role=$2, phone=$3, updated_at=NOW() 
        WHERE id=$4
    `
	_, err := r.db.Exec(query, user.Address, string(user.Role), user.Phone, id)
	return err
}

func (r *userRepository) GetAll() ([]*domain.User, error) {
	var users []*domain.User

	// เลือกดึงมาเฉพาะฟิลด์ที่ปลอดภัย (ไม่ดึง Password)
	err := r.db.Select(&users, `
		SELECT id, username, role, address, phone, is_verified, email
		FROM users
		ORDER BY id ASC
	`)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `
        SELECT id, username, password, role, address, phone, email, is_verified
        FROM users WHERE email = $1
    `
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Address, &user.Phone, &user.Email, &user.IsVerified,
	)
	return &user, err
}

func (r *userRepository) UpdateVerificationStatus(userID uint) error {
	query := `UPDATE users SET is_verified = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *userRepository) UpdatePassword(userID uint, newPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, newPassword, userID)
	return err
}
