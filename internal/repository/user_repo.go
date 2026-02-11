package repository

import (
	"simple-clothes-shop/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// 1. สร้าง User ใหม่ (ตอนสมัครสมาชิก)
func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// 2. ค้นหา User ด้วย Username (ตอน Login)
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	// ค้นหา record แรกที่ username ตรงกัน
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *userRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(id uint, user *domain.User) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(user).Error
}
