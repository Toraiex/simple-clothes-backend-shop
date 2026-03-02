// ไฟล์: internal/domain/cache.go
package domain

import (
	"context"
	"time"
)

type CacheRepository interface {
	SaveOTP(ctx context.Context, email, otp string) error
	VerifyOTP(ctx context.Context, email, inputOTP string) error
	SaveSession(ctx context.Context, refreshToken string, userID uint, duration time.Duration) error
	RevokeSession(ctx context.Context, refreshToken string) error
	GetSession(ctx context.Context, refreshToken string) (string, error)
}
