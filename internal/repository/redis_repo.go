package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// สร้าง Interface ให้ Service เรียกใช้
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

	// 1. เช็ค Cooldown (60 วิ)
	if r.rdb.Exists(ctx, cooldownKey).Val() > 0 {
		return errors.New("กรุณารอ 60 วินาทีก่อนขอรหัสใหม่")
	}

	// 2. เช็ค & บวกจำนวนครั้งที่ขอ (Atomic Increment ป้องกัน Race Condition)
	// ถ้า Key ไม่มี มันจะเริ่มที่ 0 แล้วบวกเป็น 1 อัตโนมัติ
	resendCount := r.rdb.Incr(ctx, resendKey).Val()

	if resendCount == 1 {
		// ถ้าเพิ่งขอครั้งแรก ให้ตั้งเวลาหมดอายุของโควต้าไว้ที่ 15 นาที
		r.rdb.Expire(ctx, resendKey, 15*time.Minute)
	} else if resendCount > 5 {
		return errors.New("คุณขอรหัสบ่อยเกินไป กรุณาลองใหม่ในอีก 15 นาที")
	}

	// 3. ใช้ Pipeline เพื่อเซฟข้อมูลหลายตัวพร้อมกัน (ลด Network Latency)
	pipe := r.rdb.Pipeline()
	pipe.Set(ctx, codeKey, otp, 15*time.Minute)
	pipe.Set(ctx, cooldownKey, "1", 60*time.Second) // ล็อคปุ่ม 60 วิ
	pipe.Del(ctx, attemptsKey)                      // รีเซ็ตจำนวนครั้งที่ทายผิด
	_, err := pipe.Exec(ctx)

	return err
}

// ==========================================
// 2. ระบบตรวจสอบและป้องกันการสุ่มเดา (Anti Brute-Force)
// ==========================================
func (r *cacheRepository) VerifyOTP(ctx context.Context, email, inputOTP string) error {
	codeKey := fmt.Sprintf("otp:code:%s", email)
	attemptsKey := fmt.Sprintf("otp:attempts:%s", email)

	// 1. เช็คว่ามี OTP ในระบบไหม (หรือหมดอายุไปแล้ว)
	validOTP, err := r.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return errors.New("รหัส OTP หมดอายุหรือไม่ถูกต้อง")
	} else if err != nil {
		return err
	}

	// 2. เช็คจำนวนครั้งที่ทายผิด
	attempts := r.rdb.Get(ctx, attemptsKey).Val()
	if attempts >= "3" { // ทายผิดครบ 3 ครั้ง
		r.rdb.Del(ctx, codeKey) // 💣 ทำลาย OTP ทิ้งทันที!
		r.rdb.Del(ctx, attemptsKey)
		return errors.New("คุณระบุรหัสผิดเกินจำนวนที่กำหนด รหัสนี้ถูกยกเลิกแล้ว กรุณาขอใหม่")
	}

	// 3. ตรวจสอบความถูกต้อง
	if validOTP != inputOTP {
		// ถ้าผิด ให้บวกตัวเลขการทายผิดขึ้น 1 (Atomic)
		r.rdb.Incr(ctx, attemptsKey)
		r.rdb.Expire(ctx, attemptsKey, 15*time.Minute)
		return errors.New("รหัส OTP ไม่ถูกต้อง")
	}

	// 4. ถ้าถูกต้อง ล้างขยะทิ้งได้เลย
	r.rdb.Del(ctx, codeKey, attemptsKey)
	return nil
}

// ==========================================
// 3. ระบบ Session (เบา รวดเร็ว และลบตัวเองได้)
// ==========================================
func (r *cacheRepository) SaveSession(ctx context.Context, refreshToken string, userID uint, duration time.Duration) error {
	key := fmt.Sprintf("session:%s", refreshToken)
	// เก็บค่า userID ผูกกับ Token และให้ Redis ทำลายตัวเองเมื่อหมดเวลา (duration)
	return r.rdb.Set(ctx, key, userID, duration).Err()
}

func (r *cacheRepository) RevokeSession(ctx context.Context, refreshToken string) error {
	key := fmt.Sprintf("session:%s", refreshToken)
	// ลบทิ้งออกจากระบบ (เท่ากับ Logout ถาวร)
	return r.rdb.Del(ctx, key).Err()
}

// GetSession ทำหน้าที่ดึง userID จาก Redis โดยใช้ Refresh Token เป็น Key
func (r *cacheRepository) GetSession(ctx context.Context, refreshToken string) (string, error) {
	key := fmt.Sprintf("session:%s", refreshToken)

	// ดึงค่าจาก Redis
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err // ถ้าไม่เจอจะเป็น redis.Nil (ซึ่งจะไปเป็น error ใน Service ต่อ)
	}

	return val, nil
}
