package service

import (
	"errors"
	"os"
	"time"

	"simple-clothes-shop/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

// ✅ 2. อัปเดต Constructor ให้รับ SessionRepository เข้ามาด้วย
func NewUserService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) domain.UserService {
	return &userService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (s *userService) Register(user *domain.User) error {

	// ✅ check username ซ้ำก่อน
	existing, _ := s.userRepo.GetByUsername(user.Username)
	if existing != nil {
		return errors.New("username already exists")
	}

	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return s.userRepo.Create(user)
}

func (s *userService) Login(username, password, userAgent, clientIP string) (string, string, string, error) {

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", "", "", errors.New("ไม่พบชื่อผู้ใช้งานนี้")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", "", errors.New("รหัสผ่านไม่ถูกต้อง")
	}

	// สร้าง Access Token (15 นาที)
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessToken, err := accessTokenObj.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", "", "", err
	}

	// สร้าง Refresh Token (7 วัน)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshToken, err := refreshTokenObj.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", "", err
	}

	// ✅ 4. พระเอกออกโรง: บันทึกข้อมูล Session ลง Database!
	session := &domain.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // หมดอายุพร้อม Token
	}

	err = s.sessionRepo.Create(session)
	if err != nil {
		return "", "", "", errors.New("ไม่สามารถบันทึกเซสชันได้: " + err.Error())
	}

	return accessToken, refreshToken, string(user.Role), nil
}

// ... (ฟังก์ชัน Register และ Login เหมือนเดิม) ...

// ==========================================
// 3. ดึงรายชื่อทั้งหมด (เฉพาะ Admin)
// ==========================================
func (s *userService) GetAllUsers(requesterRole domain.Role) ([]*domain.User, error) {
	if requesterRole != domain.RoleAdmin {
		return nil, errors.New("forbidden: สิทธิ์การเข้าถึงถูกปฏิเสธ เฉพาะผู้ดูแลระบบเท่านั้น")
	}
	return s.userRepo.GetAll()
}

// ==========================================
// 4. อัปเดตข้อมูลผู้ใช้งาน (ทำ Partial Update)
// ==========================================
func (s *userService) UpdateUser(requesterID uint, requesterRole domain.Role, targetID uint, input *domain.User) error {

	// 1. ตรวจสอบสิทธิ์: ต้องเป็น Admin หรือ เป็นเจ้าของบัญชีตัวเองเท่านั้น
	if requesterRole != domain.RoleAdmin && requesterID != targetID {
		return errors.New("forbidden: คุณไม่มีสิทธิ์แก้ไขข้อมูลของผู้อื่น")
	}

	// 2. FETCH: ดึงข้อมูลผู้ใช้งานเดิมจาก Database ขึ้นมาก่อน
	existingUser, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return errors.New("ไม่พบข้อมูลผู้ใช้งานนี้ในระบบ")
	}

	// 3. PATCH: อัปเดตข้อมูล "เฉพาะฟิลด์ที่มีการส่งค่ามาใหม่" (ถ้าไม่ส่งมา ให้ใช้ค่าเดิม)
	if input.Address != "" {
		existingUser.Address = input.Address
	}
	if input.Phone != "" {
		existingUser.Phone = input.Phone
	}

	// 4. ROLE LOGIC: จัดการเรื่องสิทธิ์ (Role) อย่างเข้มงวด
	if input.Role != "" {
		// ถ้าคนแก้ไม่ใช่ Admin ห้ามเปลี่ยน Role เด็ดขาด!
		if requesterRole != domain.RoleAdmin {
			return errors.New("forbidden: คุณไม่สามารถเปลี่ยนระดับสิทธิ์ (Role) ของตัวเองได้")
		}

		// ป้องกันการพิมพ์ Role มั่วๆ เข้ามา (เช่น role="hacker")
		if input.Role != domain.RoleAdmin && input.Role != domain.RoleUser {
			return errors.New("invalid role: สิทธิ์ต้องเป็น 'admin' หรือ 'user' เท่านั้น")
		}
		existingUser.Role = input.Role
	}

	// 5. SAVE: บันทึกข้อมูลที่ประกอบร่างสมบูรณ์แล้ว กลับลง Database
	return s.userRepo.Update(targetID, existingUser)
}

func (s *userService) GetUser(requesterID uint, requesterRole domain.Role, targetID uint) (*domain.User, error) {
	if requesterRole != "admin" && requesterID != targetID {
		return nil, errors.New("forbidden")
	}

	return s.userRepo.GetByID(targetID)
}

// ==========================================
// 5. ต่ออายุ Access Token (Refresh Token)
// ==========================================
func (s *userService) RefreshAccessToken(refreshToken string) (string, string, error) {
	// 1. ค้นหาและตรวจสอบเซสชันเดิม
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("unauthorized: เซสชันไม่ถูกต้อง หรือหมดอายุแล้ว")
	}
	if session.IsBlocked {
		return "", "", errors.New("unauthorized: เซสชันนี้ถูกระงับการใช้งานแล้ว")
	}
	if time.Now().After(session.ExpiresAt) {
		return "", "", errors.New("unauthorized: เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่")
	}

	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return "", "", errors.New("unauthorized: ไม่พบข้อมูลผู้ใช้งาน")
	}

	// ✅ 2. สร้าง Access Token ใบใหม่ (15 นาที)
	// (ใช้ JWT_ACCESS_SECRET ที่เราแยกไว้)
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	newAccessToken, err := accessTokenObj.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", "", err
	}

	// ✅ 3. สร้าง Refresh Token ใบใหม่! (ยืดอายุไปอีก 7 วันนับจากวันนี้)
	// (ใช้ JWT_REFRESH_SECRET ที่เราแยกไว้)
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     newExpiresAt.Unix(),
	})
	newRefreshToken, err := refreshTokenObj.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", err
	}

	// ✅ 4. สั่ง Database อัปเดต Refresh Token ทับใบเก่าทันที
	err = s.sessionRepo.UpdateRefreshToken(session.ID, newRefreshToken, newExpiresAt)
	if err != nil {
		return "", "", errors.New("ไม่สามารถอัปเดตเซสชันได้")
	}

	// คืนค่ากลับไปทั้ง 2 ใบ
	return newAccessToken, newRefreshToken, nil
}

// ==========================================
// 6. ออกจากระบบ (Logout)
// ==========================================
func (s *userService) Logout(refreshToken string) error {
	// 1. ค้นหาเซสชันจาก Refresh Token ใน Database
	session, err := s.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		// ถ้าหาไม่เจอ แสดงว่าอาจจะถูกลบหรือหมดอายุไปแล้ว ถือว่า Logout สำเร็จ
		return nil
	}

	// 2. สั่ง Block เซสชันนี้ทิ้ง (is_blocked = true)
	err = s.sessionRepo.BlockSession(session.ID)
	if err != nil {
		return errors.New("เกิดข้อผิดพลาดในการออกจากระบบ")
	}

	return nil
}
