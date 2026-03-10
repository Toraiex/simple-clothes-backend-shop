package redis

import (
	"context"
	"fmt"
	"time"

	"simple-clothes-shop/internal/domain" // 💡 1. นำเข้า domain

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus" // 💡 2. นำเข้า logrus
)

type CacheRepository interface {
	SaveOTP(ctx context.Context, email, otp string) error
	VerifyOTP(ctx context.Context, email, inputOTP string) error
	SaveSession(ctx context.Context, refreshToken string, userID uint, duration time.Duration) error
	RevokeSession(ctx context.Context, refreshToken string) error
	GetSession(ctx context.Context, refreshToken string) (string, error)
}

type cacheRepository struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) CacheRepository {
	return &cacheRepository{rdb: rdb}
}

// ==========================================
// 1. ระบบบันทึกและจำกัดการส่ง OTP (ป้องกัน Spam & Race Condition)
// ==========================================
func (r *cacheRepository) SaveOTP(ctx context.Context, email, otp string) error {
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", email)
	resendKey := fmt.Sprintf("otp:resend:%s", email)
	codeKey := fmt.Sprintf("otp:code:%s", email)
	attemptsKey := fmt.Sprintf("otp:attempts:%s", email)

	if r.rdb.Exists(ctx, cooldownKey).Val() > 0 {
		// 💡 ห่อ Error เป็น ErrTooManyRequests (429)
		return fmt.Errorf("กรุณารอ 60 วินาทีก่อนขอรหัสใหม่: %w", domain.ErrTooManyRequests)
	}

	resendCount := r.rdb.Incr(ctx, resendKey).Val()

	if resendCount == 1 {
		r.rdb.Expire(ctx, resendKey, 15*time.Minute)
	} else if resendCount > 5 {
		// 💡 ห่อ Error เป็น ErrTooManyRequests (429)
		return fmt.Errorf("คุณขอรหัสบ่อยเกินไป กรุณาลองใหม่ในอีก 15 นาที: %w", domain.ErrTooManyRequests)
	}

	pipe := r.rdb.Pipeline()
	pipe.Set(ctx, codeKey, otp, 15*time.Minute)
	pipe.Set(ctx, cooldownKey, "1", 60*time.Second)
	pipe.Del(ctx, attemptsKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		logrus.Error(err) // 🚨 ดัก Log ถ้าระบบ Redis พัง
		return err
	}

	return nil
}

// ==========================================
// 2. ระบบตรวจสอบและป้องกันการสุ่มเดา (Anti Brute-Force)
// ==========================================
func (r *cacheRepository) VerifyOTP(ctx context.Context, email, inputOTP string) error {
	codeKey := fmt.Sprintf("otp:code:%s", email)
	attemptsKey := fmt.Sprintf("otp:attempts:%s", email)

	validOTP, err := r.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		// 💡 แปลงเป็น Error มาตรฐาน (ลูกค้ากรอกช้าจนหมดเวลา)
		return fmt.Errorf("รหัส OTP หมดอายุหรือไม่ถูกต้อง: %w", domain.ErrBadParamInput)
	} else if err != nil {
		logrus.Error(err)
		return err
	}

	attempts := r.rdb.Get(ctx, attemptsKey).Val()
	if attempts >= "3" {
		r.rdb.Del(ctx, codeKey)
		r.rdb.Del(ctx, attemptsKey)
		// 💡 ห่อ Error บอกว่าโดนแบนชั่วคราว (403 Forbidden หรือ 400 ก็ได้)
		return fmt.Errorf("คุณระบุรหัสผิดเกินจำนวนที่กำหนด รหัสนี้ถูกยกเลิกแล้ว กรุณาขอใหม่: %w", domain.ErrForbidden)
	}

	if validOTP != inputOTP {
		r.rdb.Incr(ctx, attemptsKey)
		r.rdb.Expire(ctx, attemptsKey, 15*time.Minute)
		// 💡 ห่อ Error ว่าพิมพ์ผิด
		return fmt.Errorf("รหัส OTP ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
	}

	r.rdb.Del(ctx, codeKey, attemptsKey)
	return nil
}

// ==========================================
// 3. ระบบ Session (เบา รวดเร็ว และลบตัวเองได้)
// ==========================================
func (r *cacheRepository) SaveSession(ctx context.Context, refreshToken string, userID uint, duration time.Duration) error {
	key := fmt.Sprintf("session:%s", refreshToken)
	err := r.rdb.Set(ctx, key, userID, duration).Err()
	if err != nil {
		logrus.Error(err) // 🚨 ดัก Log
		return err
	}
	return nil
}

func (r *cacheRepository) RevokeSession(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("session:%s", refreshToken)
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		logrus.Error(err) // 🚨 ดัก Log
		return err
	}
	return nil
}

func (r *cacheRepository) GetSession(ctx context.Context, refreshToken string) (string, error) {
	key := fmt.Sprintf("session:%s", refreshToken)

	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// 💡 แปลงเป็น ErrNotFound แบบเงียบๆ ไม่ต้อง Log
		return "", domain.ErrNotFound
	} else if err != nil {
		logrus.Error(err) // 🚨 ดัก Log กรณี Redis ดับ
		return "", err
	}

	return val, nil
}
