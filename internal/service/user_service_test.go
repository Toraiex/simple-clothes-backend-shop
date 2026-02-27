package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/service"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// 1. สร้างตัวปลอม (Mock) สำหรับ UserRepo, CacheRepo และ EmailService
// =====================================================================

// --- 1.1 ตัวปลอมของ UserRepository ---
type mockUserRepo struct {
	mockGetByUsername    *domain.User
	mockGetByUsernameErr error
	mockCreateErr        error
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return m.mockGetByUsername, m.mockGetByUsernameErr
}
func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error          { return m.mockCreateErr }
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*domain.User, error)   { return nil, nil }
func (m *mockUserRepo) Update(ctx context.Context, id uint, user *domain.User) error { return nil }
func (m *mockUserRepo) GetAll(ctx context.Context) ([]*domain.User, error)           { return nil, nil }
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdateVerificationStatus(ctx context.Context, userID uint) error { return nil }
func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID uint, newPassword string) error {
	return nil
}

// --- 1.2 ตัวปลอมของ CacheRepository (Redis) ---
type mockCacheRepo struct{ mockSaveOTPErr error }

func (m *mockCacheRepo) SaveOTP(ctx context.Context, email string, otp string) error {
	return m.mockSaveOTPErr
}
func (m *mockCacheRepo) VerifyOTP(ctx context.Context, email string, otp string) error { return nil }
func (m *mockCacheRepo) SaveSession(ctx context.Context, token string, userID uint, duration time.Duration) error {
	return nil
}
func (m *mockCacheRepo) GetSession(ctx context.Context, token string) (string, error) { return "", nil }
func (m *mockCacheRepo) RevokeSession(ctx context.Context, token string) error        { return nil }

// --- 1.3 ตัวปลอมของ EmailService ---
type mockEmailSvc struct{}

func (m *mockEmailSvc) SendVerificationEmail(toEmail string, otpCode string) error  { return nil }
func (m *mockEmailSvc) SendPasswordResetEmail(toEmail string, otpCode string) error { return nil }

// =====================================================================
// 2. เขียนเคสทดสอบสำหรับ Register
// =====================================================================
func TestRegister(t *testing.T) {
	testUser := &domain.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}

	tests := []struct {
		name           string
		setupUserMock  func() *mockUserRepo
		setupCacheMock func() *mockCacheRepo
		expectedError  string
	}{
		{
			name: "Fail - ชื่อผู้ใช้งานซ้ำ",
			setupUserMock: func() *mockUserRepo {
				return &mockUserRepo{
					// จำลองว่าหาเจอ แปลว่าซ้ำ
					mockGetByUsername: &domain.User{ID: 1, Username: "testuser"},
				}
			},
			setupCacheMock: func() *mockCacheRepo { return &mockCacheRepo{} },
			expectedError:  "username already exists",
		},
		{
			name: "Fail - Redis ล่ม (เซฟ OTP ไม่ได้)",
			setupUserMock: func() *mockUserRepo {
				return &mockUserRepo{
					mockGetByUsername: nil, // ไม่ซ้ำ
					mockCreateErr:     nil, // DB สร้างสำเร็จ
				}
			},
			setupCacheMock: func() *mockCacheRepo {
				return &mockCacheRepo{
					mockSaveOTPErr: errors.New("redis connection refused"), // Redis พัง
				}
			},
			expectedError: "redis connection refused",
		},
		{
			name: "Success - สมัครสมาชิกสำเร็จ",
			setupUserMock: func() *mockUserRepo {
				return &mockUserRepo{
					mockGetByUsername: nil,
					mockCreateErr:     nil,
				}
			},
			setupCacheMock: func() *mockCacheRepo {
				return &mockCacheRepo{
					mockSaveOTPErr: nil, // Redis ทำงานปกติ
				}
			},
			expectedError: "", // ไม่ควรมี Error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUser := tc.setupUserMock()
			mockCache := tc.setupCacheMock()
			mockEmail := &mockEmailSvc{} // เราจำลองว่าอีเมลส่งสำเร็จเสมอ

			// ประกอบร่าง UserService
			userSvc := service.NewUserService(mockUser, mockCache, mockEmail)

			// รันคำสั่ง
			err := userSvc.Register(context.Background(), testUser)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
