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
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (s *userService) Register(user *domain.User) error {

	// ✅ check username ซ้ำก่อน
	existing, _ := s.repo.GetByUsername(user.Username)
	if existing != nil {
		return errors.New("username already exists")
	}

	if user.Role == "" {
		user.Role = domain.RoleUser
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return s.repo.Create(user)
}

// ==========================================
// 2. เข้าสู่ระบบ (Login)
// ==========================================
func (s *userService) Login(username, password string) (string, string, error) {
	// 1. ค้นหา User จาก Username
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", "", errors.New("ไม่พบชื่อผู้ใช้งานนี้")
	}

	// 2. เช็คว่ารหัสผ่านตรงกันไหม (เทียบ Hash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("รหัสผ่านไม่ถูกต้อง")

	}

	// 3. ถ้าผ่านหมด -> สร้าง JWT Token (บัตรผ่าน)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // หมดอายุใน 3 วัน

	// เซ็นชื่อกำกับด้วย Secret Key (จาก .env)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	// คืนค่า Token และ Role กลับไป
	return t, string(user.Role), nil
}
func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *userService) GetUser(requesterID uint, requesterRole domain.Role, targetID uint) (*domain.User, error) {
	if requesterRole != "admin" && requesterID != targetID {
		return nil, errors.New("forbidden")
	}

	return s.repo.GetByID(targetID)
}

// ... (ฟังก์ชัน Register และ Login เหมือนเดิม) ...

// ==========================================
// 3. ดึงรายชื่อทั้งหมด (เฉพาะ Admin)
// ==========================================
func (s *userService) GetAllUsers(requesterRole domain.Role) ([]*domain.User, error) {
	if requesterRole != domain.RoleAdmin {
		return nil, errors.New("forbidden: สิทธิ์การเข้าถึงถูกปฏิเสธ เฉพาะผู้ดูแลระบบเท่านั้น")
	}
	return s.repo.GetAll()
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
	existingUser, err := s.repo.GetByID(targetID)
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
	return s.repo.Update(targetID, existingUser)
}
