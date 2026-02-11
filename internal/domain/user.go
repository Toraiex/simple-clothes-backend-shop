package domain

import "gorm.io/gorm"

// 1. User Entity
type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
	Role     string `json:"role" gorm:"default:user"`
	Address  string `json:"address"`
}

// 2. Repository Interface
type UserRepository interface {
	GetByUsername(username string) (*User, error)
	Create(user *User) error

	GetByID(id uint) (*User, error) // 👈 เพิ่ม
	Update(id uint, user *User) error
}

// 3. Service Interface
type UserService interface {
	Register(user *User) error
	Login(username, password string) (string, string, error)

	GetUser(requesterID uint, requesterRole string, targetID uint) (*User, error)
	UpdateUser(requesterID uint, requesterRole string, targetID uint, input *User) error
}
