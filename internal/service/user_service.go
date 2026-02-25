package service

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"simple-clothes-shop/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	emailSvc    EmailService
}

// ✅ 2. อัปเดต Constructor ให้รับ SessionRepository เข้ามาด้วย
func NewUserService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, emailSvc EmailService) domain.UserService {
	return &userService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		emailSvc:    emailSvc,
	}
}

// ==========================================
// 1. ลงทะเบียน (Register)
// ==========================================
func (s *userService) Register(user *domain.User) error {
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

	// ✅ 3. สร้าง OTP และกำหนดวันหมดอายุ (เช่น 15 นาที)
	otp := generateOTP() // เรียกใช้ฟังก์ชันสุ่มตัวเลขที่เราสร้างไว้
	expiresAt := time.Now().Add(15 * time.Minute)

	user.OTPCode = otp
	user.OTPExpiresAt = &expiresAt
	user.IsVerified = false // เพิ่งสมัคร ยังไม่ได้ยืนยัน

	// 4. บันทึกลง Database
	err = s.userRepo.Create(user)
	if err != nil {
		return err
	}

	// 5. ภารกิจส่งอีเมล! (ถ้ามี Email กรอกมา)
	if user.Email != "" {
		// ✅ อัปเกรด Goroutine ให้ปลอดภัย (Safe Goroutine)
		go func(targetEmail string, targetOTP string) {
			// ดักจับ Panic ป้องกันเซิร์ฟเวอร์พัง
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("⚠️ [Recovered] เกิดข้อผิดพลาดร้ายแรงในระบบส่งอีเมล:", r)
				}
			}()

			err := s.emailSvc.SendVerificationEmail(targetEmail, targetOTP)
			if err != nil {
				fmt.Println("❌ ส่งอีเมลไม่สำเร็จ:", err)
			} else {
				fmt.Println("✅ ส่ง OTP ไปที่", targetEmail, "สำเร็จแล้ว!")
			}
		}(user.Email, otp) // 👈 ส่งค่าตัวแปรเข้าไปตรงนี้ ป้องกันการดึงค่าผิดพลาด (Closure problem)
	}

	return nil
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
func generateOTP() string {
	// สุ่มตัวเลขตั้งแต่ 100000 ถึง 999999
	return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
}

// ==========================================
// ยืนยันรหัส OTP จากอีเมล
// ==========================================
func (s *userService) VerifyEmail(email string, otp string) error {
	// 1. หาข้อมูลผู้ใช้จากอีเมล
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	// 2. เช็คว่ายืนยันไปแล้วหรือยัง
	if user.IsVerified {
		return errors.New("บัญชีนี้ได้รับการยืนยันไปแล้ว")
	}

	// 3. เช็ครหัส OTP ว่าตรงกันไหม
	if user.OTPCode != otp {
		return errors.New("รหัส OTP ไม่ถูกต้อง")
	}

	// 4. เช็คเวลาหมดอายุ (15 นาที)
	if user.OTPExpiresAt == nil || time.Now().After(*user.OTPExpiresAt) {
		return errors.New("รหัส OTP หมดอายุแล้ว กรุณาขอรหัสใหม่")
	}

	// 5. ถ้าผ่านหมดทุกด่าน ให้สั่งอัปเดต Database ได้เลย!
	return s.userRepo.UpdateVerificationStatus(user.ID)
}
func (s *userService) ResendOTP(email string) error {
	// 1. หา User
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	now := time.Now()

	// 1) cooldown 60 วิ
	if user.LastVerificationOTPSentAt != nil && now.Sub(*user.LastVerificationOTPSentAt) < 60*time.Second {
		return errors.New("คุณเพิ่งขอรหัสไปเมื่อสักครู่ กรุณารอ 60 วินาทีแล้วลองใหม่")
	}

	// 2) max resend ต่อรอบ (เช่น 5 ครั้งใน 15 นาที)
	if user.VerificationOTPResendCount >= 5 && user.OTPExpiresAt != nil && now.Before(*user.OTPExpiresAt) {
		return errors.New("คุณขอรหัสบ่อยเกินไป กรุณารอให้รหัสปัจจุบันหมดอายุก่อน")
	}

	// ผ่านแล้วค่อย generate OTP ใหม่ + อัปเดต
	newOTP := generateOTP()
	expiresAt := now.Add(15 * time.Minute)

	err = s.userRepo.UpdateOTPWithRateLimit(user.ID, newOTP, expiresAt, now, user.VerificationOTPResendCount+1)
	if err != nil {
		return err
	}

	// 4. ส่งอีเมลใหม่ (ส่งแบบเบื้องหลังเหมือนเดิม)
	go func() {
		_ = s.emailSvc.SendVerificationEmail(user.Email, newOTP)
	}()

	return nil
}

// 1. ฟังก์ชันขอรีเซ็ตรหัสผ่าน
func (s *userService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// ✅ พิมพ์ Error จริงออกมาดูใน Terminal ของเราด้วย
		fmt.Println("❌ GetByEmail Error:", err)
		return errors.New("ไม่พบอีเมลนี้ในระบบ หรือเกิดข้อผิดพลาดภายใน")
	}

	otp := generateOTP() // ใช้ฟังก์ชันสุ่ม OTP เดิมที่มีอยู่แล้ว
	expiresAt := time.Now().Add(15 * time.Minute)
	now := time.Now()

	// อัปเดต OTP สำหรับ reset password พร้อมรีเซ็ตข้อมูล rate limit ฝั่ง reset
	err = s.userRepo.UpdateResetOTPWithRateLimit(user.ID, otp, expiresAt, now, 0)
	if err != nil {
		return err
	}

	// ส่งอีเมลเบื้องหลัง
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
func (s *userService) ResetPassword(email, otp, newPassword string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	if user.OTPCode != otp || user.OTPExpiresAt == nil || time.Now().After(*user.OTPExpiresAt) {
		return errors.New("รหัส OTP ไม่ถูกต้องหรือหมดอายุแล้ว")
	}

	// เข้ารหัสผ่านใหม่
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}

	// อัปเดตรหัสผ่านลง DB (ต้องไปเพิ่มท่า UpdatePassword ใน Repository นิดนึง)
	err = s.userRepo.UpdatePassword(user.ID, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

// 3. ฟังก์ชันขอส่ง OTP สำหรับรีเซ็ตรหัสผ่านซ้ำ (Resend Reset OTP)
func (s *userService) ResendResetOTP(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	now := time.Now()

	// 1) cooldown 60 วิ สำหรับ reset OTP
	if user.LastResetOTPSentAt != nil && now.Sub(*user.LastResetOTPSentAt) < 60*time.Second {
		return errors.New("คุณเพิ่งขอรหัสรีเซ็ตรหัสผ่านไปเมื่อสักครู่ กรุณารอ 60 วินาทีแล้วลองใหม่")
	}

	// 2) max resend ต่อรอบ (เช่น 5 ครั้งใน 15 นาที) สำหรับ reset OTP
	if user.ResetOTPResendCount >= 5 && user.OTPExpiresAt != nil && now.Before(*user.OTPExpiresAt) {
		return errors.New("คุณขอรหัสรีเซ็ตรหัสผ่านบ่อยเกินไป กรุณารอให้รหัสปัจจุบันหมดอายุก่อน")
	}

	newOTP := generateOTP()
	expiresAt := now.Add(15 * time.Minute)

	err = s.userRepo.UpdateResetOTPWithRateLimit(user.ID, newOTP, expiresAt, now, user.ResetOTPResendCount+1)
	if err != nil {
		return err
	}

	go func() {
		_ = s.emailSvc.SendPasswordResetEmail(user.Email, newOTP)
	}()

	return nil
}
