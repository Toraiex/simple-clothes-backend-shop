package domain

import (
	"context"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// 1. Struct ของผู้ใช้งาน (สะอาดหมดจด ไม่มี OTP แล้ว)
type User struct {
	ID         uint   `db:"id" json:"id"`
	Username   string `db:"username" json:"username"`
	Password   string `db:"password" json:"-"` // ซ่อนรหัสผ่านไม่ให้หลุดไปกับ API
	Role       Role   `db:"role" json:"role"`
	Address    string `db:"address" json:"address"`
	Phone      string `db:"phone" json:"phone"`
	Email      string `db:"email" json:"email"`
	IsVerified bool   `db:"is_verified" json:"is_verified"`
}

// 2. Repository Interface (เชื่อม ctx ทุกตัว)
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, id uint, user *User) error
	GetAll(ctx context.Context) ([]*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)

	UpdateVerificationStatus(ctx context.Context, userID uint) error
	UpdatePassword(ctx context.Context, userID uint, newPassword string) error
}

// 3. Service Interface (เชื่อม ctx ทุกตัว)
type UserUsecase interface {
	Register(ctx context.Context, user *User) error
	Login(ctx context.Context, username, password string) (string, string, string, error)
	GetUser(ctx context.Context, requesterID uint, requesterRole Role, targetID uint) (*User, error)
	UpdateUser(ctx context.Context, requesterID uint, requesterRole Role, targetID uint, input *User) error
	GetAllUsers(ctx context.Context, requesterRole Role) ([]*User, error)

	RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error

	VerifyEmail(ctx context.Context, email string, otp string) error
	ResendOTP(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
	ResendResetOTP(ctx context.Context, email string) error
}
