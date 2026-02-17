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
func (s *userService) UpdateUser(requesterID uint, requesterRole domain.Role, targetID uint, input *domain.User) error {

	if requesterRole != domain.RoleAdmin && requesterID != targetID {
		return errors.New("forbidden")
	}

	if requesterRole != domain.RoleAdmin {
		input.Role = ""
	}

	return s.repo.Update(targetID, input)
}
