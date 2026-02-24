package repository

import (
	"simple-clothes-shop/internal/domain"
	"time"

	"github.com/jmoiron/sqlx"
)

// สร้าง Struct สำหรับ Session Repository
type sessionRepository struct {
	db *sqlx.DB
}

// ฟังก์ชันสำหรับสร้าง (Instantiate) Repository ตัวนี้ขึ้นมาใช้งาน
func NewSessionRepository(db *sqlx.DB) domain.SessionRepository {
	return &sessionRepository{db: db}
}

// ==========================================
// 1. บันทึกเซสชันใหม่ (ตอน Login)
// ==========================================
func (r *sessionRepository) Create(session *domain.Session) error {
	// 💡 ข้อสังเกตระดับโปร: เราไม่ส่ง ID ไปบันทึก
	// เพราะเราตั้งให้ PostgreSQL สร้าง UUID ให้เอง (gen_random_uuid())
	// แต่เราใช้คำสั่ง RETURNING id เพื่อดึงค่า UUID นั้นกลับมาเก็บใส่ Struct ให้เอาไปใช้ต่อได้
	return r.db.QueryRow(`
		INSERT INTO sessions (user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.ClientIP,
		session.IsBlocked,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

// ==========================================
// 2. ดึงข้อมูลเซสชัน (ตอนขอ Access Token ใหม่ / ตรวจสอบสิทธิ์)
// ==========================================
func (r *sessionRepository) GetByID(id string) (*domain.Session, error) {
	var session domain.Session

	// 💡 ใช้ r.db.Get ของ sqlx เพื่อดึงข้อมูลแถวเดียวมาใส่ Struct อัตโนมัติ
	err := r.db.Get(&session, `
		SELECT id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
		FROM sessions
		WHERE id=$1
	`, id)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) BlockSession(id string) error {
	// 💡 ทำแค่การ Update ค่า is_blocked เป็น true (Soft Delete / Revoke)
	// ดีกว่าการ DELETE ทิ้งไปเลย เพราะเรายังเก็บประวัติไว้ดูย้อนหลังได้ว่าเคยมีเครื่องไหนล็อกอินบ้าง
	_, err := r.db.Exec(`
		UPDATE sessions
		SET is_blocked = true
		WHERE id=$1
	`, id)

	return err
}

// ==========================================
// 4. ค้นหาเซสชันด้วย Refresh Token
// ==========================================
func (r *sessionRepository) GetByRefreshToken(token string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.Get(&session, `
		SELECT id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
		FROM sessions
		WHERE refresh_token=$1
	`, token)

	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ==========================================
// 5. อัปเดต Refresh Token ใบใหม่ (Rotation)
// ==========================================
func (r *sessionRepository) UpdateRefreshToken(sessionID string, newRefreshToken string, newExpiresAt time.Time) error {
	// 💡 สั่ง UPDATE ทับข้อมูลเดิมในแถวเดิม ช่วยประหยัดพื้นที่ Database
	_, err := r.db.Exec(`
		UPDATE sessions
		SET refresh_token = $1, expires_at = $2
		WHERE id = $3
	`, newRefreshToken, newExpiresAt, sessionID)

	return err
}
