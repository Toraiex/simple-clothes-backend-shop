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

func (r *userRepository) Create(user *domain.User) error {

	return r.db.QueryRow(`
		INSERT INTO users (username, password, role, address)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`,
		user.Username,
		user.Password,
		string(user.Role),
		user.Address,
	).Scan(&user.ID)
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {

	var user domain.User

	err := r.db.Get(&user, `
		SELECT id, username, password, role, address
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
		SELECT id, username, password, role, address
		FROM users
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(id uint, user *domain.User) error {

	_, err := r.db.Exec(`
		UPDATE users
		SET address=$1, role=$2
		WHERE id=$3
	`,
		user.Address,
		string(user.Role),
		id,
	)

	return err
}
