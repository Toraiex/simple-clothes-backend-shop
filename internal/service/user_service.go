package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo         domain.UserRepository
	cacheRepo        repository.CacheRepository
	emailSvc         EmailService
	jwtAccessSecret  string
	jwtRefreshSecret string
}

func NewUserService(userRepo domain.UserRepository, cacheRepo repository.CacheRepository, emailSvc EmailService) domain.UserService {
	return &userService{
		userRepo:         userRepo,
		cacheRepo:        cacheRepo,
		emailSvc:         emailSvc,
		jwtAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		jwtRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
	}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (s *userService) Register(ctx context.Context, user *domain.User) error {
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
	user.IsVerified = false // 👈 ไม่ต้องเซ็ตค่า OTP ลง Struct นี้แล้ว เพราะเราจะเก็บใน Redis

	// 1. บันทึกลง PostgreSQL (ข้อมูลถาวร)
	err = s.userRepo.Create(user)
	if err != nil {
		return err
	}

	// 2. ถ้ามีอีเมล ให้สร้าง OTP และเก็บลง Redis (ข้อมูลชั่วคราว)
	if user.Email != "" {
		otp := generateOTP()

		// 🚀 โยนให้ Redis จัดการ Rate Limit และวันหมดอายุ
		err = s.cacheRepo.SaveOTP(ctx, user.Email, otp)
		if err != nil {
			return err // กรณีที่ผู้ใช้สมัครรัวๆ Redis จะเตะกลับตรงนี้
		}

		// 3. ภารกิจส่งอีเมล (Safe Goroutine)
		go func(targetEmail string, targetOTP string) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("⚠️ [Recovered] Email sending panic:", r)
				}
			}()
			_ = s.emailSvc.SendVerificationEmail(targetEmail, targetOTP)
		}(user.Email, otp)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, string, string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", "", "", errors.New("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง") // 🛡️ เปลี่ยน Error ให้คลุมเครือ กันการเดา Username
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", "", errors.New("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง")
	}

	// สร้าง Access Token (15 นาที)
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		return "", "", "", err
	}

	// สร้าง Refresh Token (7 วัน)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		return "", "", "", err
	}

	// 🚀 บันทึก Session ลง Redis (ลบตัวเองทิ้งเมื่อครบ 7 วัน)
	err = s.cacheRepo.SaveSession(ctx, refreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", "", errors.New("ไม่สามารถสร้างเซสชันได้")
	}

	return accessToken, refreshToken, string(user.Role), nil
}

func (s *userService) GetAllUsers(requesterRole domain.Role) ([]*domain.User, error) {
	if requesterRole != domain.RoleAdmin {
		return nil, errors.New("forbidden: สิทธิ์การเข้าถึงถูกปฏิเสธ เฉพาะผู้ดูแลระบบเท่านั้น")
	}
	return s.userRepo.GetAll()
}

func (s *userService) UpdateUser(requesterID uint, requesterRole domain.Role, targetID uint, input *domain.User) error {
	current, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return errors.New("ไม่พบข้อมูลผู้ใช้งานนี้ในระบบ")
	}

	if requesterRole != domain.RoleAdmin && requesterID != targetID {
		return errors.New("forbidden: คุณไม่มีสิทธิ์แก้ไขข้อมูลของผู้อื่น")
	}

	if input.Address != "" {
		current.Address = input.Address
	}
	if input.Phone != "" {
		current.Phone = input.Phone
	}

	if input.Role != "" && requesterRole == domain.RoleAdmin {
		if input.Role == domain.RoleAdmin || input.Role == domain.RoleUser {
			current.Role = input.Role
		}
	}

	return s.userRepo.Update(targetID, current)
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
// ==========================================
// 5. ต่ออายุ Access Token (Refresh Token)
// ==========================================
func (s *userService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 1. ตรวจสอบว่า Refresh Token นี้มีอยู่ใน Redis หรือไม่
	// (สมมติว่าคุณเพิ่มฟังก์ชัน GetSession ใน CacheRepository แล้ว)
	userIDStr, err := s.cacheRepo.GetSession(ctx, refreshToken)
	if err != nil {
		return "", "", errors.New("unauthorized: เซสชันไม่ถูกต้อง หรือหมดอายุแล้ว")
	}

	// แปลง userIDStr กลับเป็น uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return "", "", errors.New("unauthorized: ข้อมูลเซสชันผิดพลาด")
	}

	user, err := s.userRepo.GetByID(uint(userID))
	if err != nil {
		return "", "", errors.New("unauthorized: ไม่พบข้อมูลผู้ใช้งาน")
	}

	// 2. สร้าง Access Token ใบใหม่ (15 นาที)
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	newAccessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		return "", "", err
	}

	// 3. สร้าง Refresh Token ใบใหม่! (ยืดอายุไปอีก 7 วันนับจากวันนี้)
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     newExpiresAt.Unix(),
	})
	newRefreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret)) // 👈 แก้ตรงนี้ให้ใช้ตัวแปร struct
	if err != nil {
		return "", "", err
	}

	// 4. สั่ง Redis สลับ Token ทันที
	// - ลบ Token เก่าทิ้ง
	_ = s.cacheRepo.RevokeSession(ctx, refreshToken)
	// - สร้าง Token ใหม่
	err = s.cacheRepo.SaveSession(ctx, newRefreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", errors.New("ไม่สามารถอัปเดตเซสชันได้")
	}

	return newAccessToken, newRefreshToken, nil
}

// ==========================================
// 6. ออกจากระบบ (Logout)
// ==========================================
func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	// 🚀 ลบ Key ออกจาก Redis (Revoke Session อย่างสมบูรณ์แบบและรวดเร็ว)
	return s.cacheRepo.RevokeSession(ctx, refreshToken)
}

// ==========================================
// 8. สร้างเลขสุ่ม 6 หลัก
// ==========================================
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
}

// ==========================================
// ยืนยันรหัส OTP จากอีเมล
// ==========================================
func (s *userService) VerifyEmail(ctx context.Context, email string, otp string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	if user.IsVerified {
		return errors.New("บัญชีนี้ได้รับการยืนยันไปแล้ว")
	}

	// 🚀 ให้ Redis ตรวจสอบ OTP (เช็คเวลาหมดอายุ + ดักจับ Brute-force 3 ครั้ง)
	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return err // จะคืนค่า error จาก Redis เช่น "รหัสผิด", "หมดอายุ", หรือ "ทายผิดเกินบล็อก"
	}

	// อัปเดต PostgreSQL ว่ายืนยันแล้ว
	return s.userRepo.UpdateVerificationStatus(user.ID)
}

func (s *userService) ResendOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	newOTP := generateOTP()

	// 🚀 Redis ตรวจสอบ Cooldown (60 วิ) และ Max Resend ให้อัตโนมัติ!
	// (ลอจิกเช็คเวลาเดิมๆ ที่รกๆ ลบทิ้งไปได้เลย)
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		return err // ถ้าติด Cooldown หรือเกิน Limit Redis จะส่ง Error กลับมาเอง
	}

	go func() {
		_ = s.emailSvc.SendVerificationEmail(user.Email, newOTP)
	}()

	return nil
}

// 1. ฟังก์ชันขอรีเซ็ตรหัสผ่าน
func (s *userService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// 🛡️ [Security] ไม่บอกว่า "ไม่พบอีเมล" เพื่อป้องกันแฮกเกอร์สุ่มหาอีเมลผู้ใช้งาน
		fmt.Println("⚠️ พยายามขอรีเซ็ตรหัสผ่านแต่อีเมลไม่มีในระบบ:", email)
		return nil // 👈 ตอบว่าสำเร็จ (แต่จริงๆ ไม่ได้ส่งอะไร)
	}

	otp := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, otp)
	if err != nil {
		return err
	}

	go func(targetEmail, targetOTP string) {
		defer func() {
			if r := recover(); r != nil {
			}
		}()
		_ = s.emailSvc.SendPasswordResetEmail(targetEmail, targetOTP)
	}(user.Email, otp)

	return nil
}

// 2. ฟังก์ชันตั้งรหัสผ่านใหม่
func (s *userService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("คำขอไม่ถูกต้อง") // 🛡️ คลุมเครือไว้ก่อน
	}

	// 🚀 ให้ Redis ยืนยัน OTP ป้องกันการสุ่มเดารหัสแบบ Brute-force
	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user.ID, string(hashedPassword))
}

// 3. ฟังก์ชันขอส่ง OTP สำหรับรีเซ็ตรหัสผ่านซ้ำ (Resend Reset OTP)
func (s *userService) ResendResetOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	newOTP := generateOTP()

	// 🚀 โยนให้ Redis ตรวจสอบ Cooldown และ Limit อัตโนมัติ
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		return err // เตะกลับถ้าติด Cooldown หรือเกินโควต้า
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("⚠️ [Recovered] Email sending panic in ResendResetOTP:", r)
			}
		}()
		_ = s.emailSvc.SendPasswordResetEmail(user.Email, newOTP)
	}()

	return nil
}
