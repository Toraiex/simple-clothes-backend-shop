package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"simple-clothes-shop/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus" // 💡 เพิ่ม logrus
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo         domain.UserRepository
	cacheRepo        domain.CacheRepository
	emailSvc         domain.EmailService
	jwtAccessSecret  string
	jwtRefreshSecret string
}

func NewUserService(userRepo domain.UserRepository, cacheRepo domain.CacheRepository, emailSvc domain.EmailService) domain.UserUsecase {
	return &userService{
		userRepo:         userRepo,
		cacheRepo:        cacheRepo,
		emailSvc:         emailSvc,
		jwtAccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
		jwtRefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
	}
}

func (s *userService) Register(ctx context.Context, user *domain.User) error {
	_, err := s.userRepo.GetByUsername(ctx, user.Username)
	if err == nil {
		return fmt.Errorf("ชื่อผู้ใช้งาน '%s' มีอยู่ในระบบแล้ว: %w", user.Username, domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		logrus.Error(err)
		return domain.ErrInternalServerError
	}
	user.Password = string(hashedPassword)
	user.IsVerified = false

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	if user.Email != "" {
		otp := generateOTP()
		err = s.cacheRepo.SaveOTP(ctx, user.Email, otp)
		if err != nil {
			logrus.Error(err)
			return domain.ErrInternalServerError
		}

		go func(targetEmail string, targetOTP string) {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("[Recovered] Email sending panic: %v", r)
				}
			}()
			_ = s.emailSvc.SendVerificationEmail(targetEmail, targetOTP)
		}(user.Email, otp)
	}

	return nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", "", fmt.Errorf("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง: %w", domain.ErrUnauthorized)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", "", fmt.Errorf("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง: %w", domain.ErrUnauthorized)
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		logrus.Error(err)
		return "", "", "", domain.ErrInternalServerError
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		logrus.Error(err)
		return "", "", "", domain.ErrInternalServerError
	}

	err = s.cacheRepo.SaveSession(ctx, refreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		logrus.Error(err)
		return "", "", "", fmt.Errorf("ไม่สามารถสร้างเซสชันได้: %w", domain.ErrInternalServerError)
	}

	return accessToken, refreshToken, string(user.Role), nil
}

func (s *userService) GetAllUsers(ctx context.Context, requesterRole domain.Role) ([]*domain.User, error) {
	if requesterRole != domain.RoleAdmin {
		return nil, fmt.Errorf("สิทธิ์การเข้าถึงถูกปฏิเสธ: %w", domain.ErrForbidden)
	}
	return s.userRepo.GetAll(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, requesterID uint, requesterRole domain.Role, targetID uint, input *domain.User) error {
	current, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("ไม่พบข้อมูลผู้ใช้งานนี้ในระบบ: %w", domain.ErrNotFound)
	}

	if requesterRole != domain.RoleAdmin && requesterID != targetID {
		return fmt.Errorf("คุณไม่มีสิทธิ์แก้ไขข้อมูลของผู้อื่น: %w", domain.ErrForbidden)
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

	return s.userRepo.Update(ctx, targetID, current)
}

func (s *userService) GetUser(ctx context.Context, requesterID uint, requesterRole domain.Role, targetID uint) (*domain.User, error) {
	if requesterRole != "admin" && requesterID != targetID {
		return nil, fmt.Errorf("ไม่มีสิทธิ์เข้าถึง: %w", domain.ErrForbidden)
	}
	return s.userRepo.GetByID(ctx, targetID)
}

func (s *userService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	userIDStr, err := s.cacheRepo.GetSession(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("เซสชันไม่ถูกต้อง หรือหมดอายุแล้ว: %w", domain.ErrUnauthorized)
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		logrus.Error(err)
		return "", "", fmt.Errorf("ข้อมูลเซสชันผิดพลาด: %w", domain.ErrUnauthorized)
	}

	user, err := s.userRepo.GetByID(ctx, uint(userID))
	if err != nil {
		return "", "", fmt.Errorf("ไม่พบข้อมูลผู้ใช้งาน: %w", domain.ErrUnauthorized)
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	newAccessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		logrus.Error(err)
		return "", "", domain.ErrInternalServerError
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     newExpiresAt.Unix(),
	})
	newRefreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		logrus.Error(err)
		return "", "", domain.ErrInternalServerError
	}

	_ = s.cacheRepo.RevokeSession(ctx, refreshToken)
	err = s.cacheRepo.SaveSession(ctx, newRefreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		logrus.Error(err)
		return "", "", domain.ErrInternalServerError
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	return s.cacheRepo.RevokeSession(ctx, refreshToken)
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
}

func (s *userService) VerifyEmail(ctx context.Context, email string, otp string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("ไม่พบอีเมลนี้ในระบบ: %w", domain.ErrNotFound)
	}

	if user.IsVerified {
		return fmt.Errorf("บัญชีนี้ได้รับการยืนยันไปแล้ว: %w", domain.ErrConflict)
	}

	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return fmt.Errorf("รหัส OTP ไม่ถูกต้อง: %w", domain.ErrBadParamInput) // ถือว่าลูกค้ากรอกผิด
	}

	return s.userRepo.UpdateVerificationStatus(ctx, user.ID)
}

func (s *userService) ResendOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("ไม่พบอีเมลนี้ในระบบ: %w", domain.ErrNotFound)
	}
	newOTP := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		if errors.Is(err, domain.ErrTooManyRequests) {
			return err
		}
		return domain.ErrInternalServerError
	}

	go func() {
		_ = s.emailSvc.SendVerificationEmail(user.Email, newOTP)
	}()

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logrus.Warnf("พยายามขอรีเซ็ตรหัสผ่านแต่อีเมลไม่มีในระบบ: %s", email)
		return nil
	}

	otp := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, otp)
	if err != nil {
		logrus.Error(err)
		return domain.ErrInternalServerError
	}

	go func(targetEmail, targetOTP string) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("[Recovered] Email sending panic: %v", r)
			}
		}()
		_ = s.emailSvc.SendPasswordResetEmail(targetEmail, targetOTP)
	}(user.Email, otp)

	return nil
}

func (s *userService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("คำขอไม่ถูกต้อง: %w", domain.ErrNotFound)
	}

	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return fmt.Errorf("รหัส OTP ไม่ถูกต้องหรือหมดอายุ: %w", domain.ErrBadParamInput)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		logrus.Error(err)
		return domain.ErrInternalServerError
	}

	return s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
}

func (s *userService) ResendResetOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("ไม่พบอีเมลนี้ในระบบ: %w", domain.ErrNotFound)
	}

	newOTP := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		logrus.Error(err)
		return domain.ErrInternalServerError
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("[Recovered] Email sending panic in ResendResetOTP: %v", r)
			}
		}()
		_ = s.emailSvc.SendPasswordResetEmail(user.Email, newOTP)
	}()

	return nil
}
