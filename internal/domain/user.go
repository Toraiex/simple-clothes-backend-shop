package domain

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password" json:"-"` // เพิ่ม json:"-" เพื่อป้องกันรหัสผ่านหลุดตอนส่ง Response อัตโนมัติ
	Role     Role   `db:"role"`
	Address  string `db:"address"`
	Phone    string `db:"phone" json:"phone"`

	Email        string     `db:"email" json:"email"`
	IsVerified   bool       `db:"is_verified" json:"is_verified"`
	OTPCode      string     `db:"otp_code" json:"-"`
	OTPExpiresAt *time.Time `db:"otp_expires_at" json:"-"`
}

// โครงสร้างข้อมูลให้ตรงกับตาราง sessions
type Session struct {
	ID           string    `db:"id"`
	UserID       uint      `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	UserAgent    string    `db:"user_agent"`
	ClientIP     string    `db:"client_ip"`
	IsBlocked    bool      `db:"is_blocked"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

// 2. Repository Interface
type UserRepository interface {
	GetByUsername(username string) (*User, error)
	Create(user *User) error
	GetByID(id uint) (*User, error)
	Update(id uint, user *User) error
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	UpdateVerificationStatus(userID uint) error
	UpdateOTP(userID uint, otp string, expiresAt time.Time) error
}

// 3. Service Interface
type UserService interface {
	Register(user *User) error
	Login(username, password, userAgent, clientIP string) (string, string, string, error)
	GetUser(requesterID uint, requesterRole Role, targetID uint) (*User, error)
	UpdateUser(requesterID uint, requesterRole Role, targetID uint, input *User) error
	GetAllUsers(requesterRole Role) ([]*User, error)
	RefreshAccessToken(refreshToken string) (string, string, error)
	Logout(refreshToken string) error
	VerifyEmail(email string, otp string) error
	ResendOTP(email string) error
}

type SessionRepository interface {
	Create(session *Session) error
	GetByID(id string) (*Session, error)
	BlockSession(id string) error
	GetByRefreshToken(token string) (*Session, error)
	UpdateRefreshToken(sessionID string, newRefreshToken string, newExpiresAt time.Time) error
}
