package repository

import (
	"simple-clothes-shop/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}
type UserModel struct {
	gorm.Model
	Username string
	Password string
	Role     string
	Address  string
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// 1. สร้าง User ใหม่ (ตอนสมัครสมาชิก)
func (r *userRepository) Create(user *domain.User) error {
	model := UserModel{
		Username: user.Username,
		Password: user.Password,
		Role:     string(user.Role),
		Address:  user.Address,
	}

	return r.db.Create(&model).Error
}

// 2. ค้นหา User ด้วย Username (ตอน Login)
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var model UserModel

	if err := r.db.Where("username = ?", username).First(&model).Error; err != nil {
		return nil, err
	}

	return &domain.User{
		ID:       model.ID,
		Username: model.Username,
		Password: model.Password,
		Role:     domain.Role(model.Role),
		Address:  model.Address,
	}, nil
}
func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var model UserModel

	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}

	return &domain.User{
		ID:       model.ID,
		Username: model.Username,
		Password: model.Password,
		Role:     domain.Role(model.Role),
		Address:  model.Address,
	}, nil
}

func (r *userRepository) Update(id uint, user *domain.User) error {
	updateData := map[string]interface{}{}

	if user.Address != "" {
		updateData["address"] = user.Address
	}

	if user.Role != "" {
		updateData["role"] = string(user.Role)
	}

	return r.db.Model(&UserModel{}).
		Where("id = ?", id).
		Updates(updateData).Error
}
