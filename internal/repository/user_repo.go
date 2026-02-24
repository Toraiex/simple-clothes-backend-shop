package repository

import (
	"simple-clothes-shop/internal/domain"
	"time"

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
	// ✅ อัปเดตคำสั่ง SQL ให้รองรับ Email, IsVerified, OTP และวันหมดอายุ
	query := `
		INSERT INTO users (
			username, password, role, address, phone, 
			email, is_verified, otp_code, otp_expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRow(query,
		user.Username,
		user.Password,
		user.Role,
		user.Address,
		user.Phone,
		user.Email,        // 👈 เพิ่มเข้ามา
		user.IsVerified,   // 👈 เพิ่มเข้ามา
		user.OTPCode,      // 👈 เพิ่มเข้ามา
		user.OTPExpiresAt, // 👈 เพิ่มเข้ามา
	).Scan(&user.ID)

	return err
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {

	var user domain.User

	err := r.db.Get(&user, `
		SELECT id, username, password, role, address, phone, email
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
		SELECT id, username, password, role, address, phone, email
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
		SET address=$1, role=$2, phone=$3
		WHERE id=$4
	`,
		user.Address,
		string(user.Role),
		user.Phone,
		id,
	)

	return err
}

func (r *userRepository) GetAll() ([]*domain.User, error) {
	var users []*domain.User

	// เลือกดึงมาเฉพาะฟิลด์ที่ปลอดภัย (ไม่ดึง Password)
	err := r.db.Select(&users, `
		SELECT id, username, role, address, phone, email
		FROM users
		ORDER BY id ASC
	`)

	if err != nil {
		return nil, err
	}

	return users, nil
}

// ==========================================
// ค้นหาผู้ใช้งานจาก Email
// ==========================================
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	// 💡 ตรวจสอบตรงนี้: ต้อง SELECT ให้ครบทุกคอลัมน์ตามลำดับที่ Scan ด้านล่างครับ
	query := `SELECT id, username, password, role, address, phone, email, is_verified, otp_code, otp_expires_at FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.Address, &user.Phone, &user.Email, &user.IsVerified,
		&user.OTPCode, &user.OTPExpiresAt,
	)
	return &user, err
}

// ==========================================
// อัปเดตสถานะว่ายืนยันอีเมลแล้ว (และลบ OTP ทิ้ง)
// ==========================================
func (r *userRepository) UpdateVerificationStatus(userID uint) error {
	query := `
		UPDATE users 
		SET is_verified = true, otp_code = NULL, otp_expires_at = NULL 
		WHERE id = $1
	`
	_, err := r.db.Exec(query, userID)
	return err
}
func (r *userRepository) UpdateOTP(userID uint, otp string, expiresAt time.Time) error {
	query := `UPDATE users SET otp_code = $1, otp_expires_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, otp, expiresAt, userID)
	return err
}
