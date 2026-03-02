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
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo         domain.UserRepository
	cacheRepo        domain.CacheRepository // ✅ เติม domain. นำหน้า
	emailSvc         domain.EmailService    // ✅ เติม domain. นำหน้า
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
	existing, _ := s.userRepo.GetByUsername(ctx, user.Username)
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
	user.IsVerified = false

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	if user.Email != "" {
		otp := generateOTP()
		err = s.cacheRepo.SaveOTP(ctx, user.Email, otp)
		if err != nil {
			return err
		}

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
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", "", errors.New("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", "", errors.New("ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง")
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	accessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		return "", "", "", err
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		return "", "", "", err
	}

	err = s.cacheRepo.SaveSession(ctx, refreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", "", errors.New("ไม่สามารถสร้างเซสชันได้")
	}

	return accessToken, refreshToken, string(user.Role), nil
}

func (s *userService) GetAllUsers(ctx context.Context, requesterRole domain.Role) ([]*domain.User, error) {
	if requesterRole != domain.RoleAdmin {
		return nil, errors.New("forbidden: สิทธิ์การเข้าถึงถูกปฏิเสธ เฉพาะผู้ดูแลระบบเท่านั้น")
	}
	return s.userRepo.GetAll(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, requesterID uint, requesterRole domain.Role, targetID uint, input *domain.User) error {
	current, err := s.userRepo.GetByID(ctx, targetID)
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

	return s.userRepo.Update(ctx, targetID, current)
}

func (s *userService) GetUser(ctx context.Context, requesterID uint, requesterRole domain.Role, targetID uint) (*domain.User, error) {
	if requesterRole != "admin" && requesterID != targetID {
		return nil, errors.New("forbidden")
	}
	return s.userRepo.GetByID(ctx, targetID)
}

func (s *userService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	userIDStr, err := s.cacheRepo.GetSession(ctx, refreshToken)
	if err != nil {
		return "", "", errors.New("unauthorized: เซสชันไม่ถูกต้อง หรือหมดอายุแล้ว")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return "", "", errors.New("unauthorized: ข้อมูลเซสชันผิดพลาด")
	}

	user, err := s.userRepo.GetByID(ctx, uint(userID))
	if err != nil {
		return "", "", errors.New("unauthorized: ไม่พบข้อมูลผู้ใช้งาน")
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	newAccessToken, err := accessTokenObj.SignedString([]byte(s.jwtAccessSecret))
	if err != nil {
		return "", "", err
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     newExpiresAt.Unix(),
	})
	newRefreshToken, err := refreshTokenObj.SignedString([]byte(s.jwtRefreshSecret))
	if err != nil {
		return "", "", err
	}

	_ = s.cacheRepo.RevokeSession(ctx, refreshToken)
	err = s.cacheRepo.SaveSession(ctx, newRefreshToken, user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", errors.New("ไม่สามารถอัปเดตเซสชันได้")
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
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	if user.IsVerified {
		return errors.New("บัญชีนี้ได้รับการยืนยันไปแล้ว")
	}

	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateVerificationStatus(ctx, user.ID)
}

func (s *userService) ResendOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	newOTP := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		return err
	}

	go func() {
		_ = s.emailSvc.SendVerificationEmail(user.Email, newOTP)
	}()

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		fmt.Println("⚠️ พยายามขอรีเซ็ตรหัสผ่านแต่อีเมลไม่มีในระบบ:", email)
		return nil
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

func (s *userService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("คำขอไม่ถูกต้อง")
	}

	err = s.cacheRepo.VerifyOTP(ctx, email, otp)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
}

func (s *userService) ResendResetOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("ไม่พบอีเมลนี้ในระบบ")
	}

	newOTP := generateOTP()
	err = s.cacheRepo.SaveOTP(ctx, email, newOTP)
	if err != nil {
		return err
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
