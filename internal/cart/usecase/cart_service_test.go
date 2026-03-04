package usecase_test

import (
	"context"
	"errors"
	"testing"

	"simple-clothes-shop/internal/cart/usecase"
	"simple-clothes-shop/internal/domain"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// 1. สร้างตัวปลอม (Mock) สำหรับ CartRepository
// =====================================================================
type mockCartRepo struct {
	// ตัวแปรสำหรับควบคุมผลลัพธ์ว่าอยากให้ DB ปลอมตอบกลับมาว่าอะไร
	mockGetCart   *domain.Cart
	mockGetErr    error
	mockCreate    *domain.Cart
	mockCreateErr error
	mockAddErr    error
}

func (m *mockCartRepo) GetCartByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	return m.mockGetCart, m.mockGetErr
}
func (m *mockCartRepo) CreateCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	return m.mockCreate, m.mockCreateErr
}
func (m *mockCartRepo) AddItem(ctx context.Context, cartID uint, variantID uint, quantity int) error {
	return m.mockAddErr
}
func (m *mockCartRepo) GetCartItemsWithDetails(ctx context.Context, cartID uint) ([]domain.CartItem, error) {
	return []domain.CartItem{}, nil
}
func (m *mockCartRepo) UpdateItemQuantity(ctx context.Context, cartItemID uint, quantity int) error {
	return nil
}
func (m *mockCartRepo) RemoveItem(ctx context.Context, cartItemID uint) error { return nil }
func (m *mockCartRepo) ClearCart(ctx context.Context, cartID uint) error      { return nil }

// =====================================================================
// 2. เริ่มเขียนเคสทดสอบสำหรับ AddToCart
// =====================================================================
func TestAddToCart(t *testing.T) {
	tests := []struct {
		name          string // ชื่อสถานการณ์
		inputQuantity int    // จำนวนที่ลูกค้ากดสั่ง

		// ตั้งค่าพฤติกรรมของ DB ปลอม
		setupMockRepo func() *mockCartRepo

		expectedError string // ผลลัพธ์ที่คาดหวัง
	}{
		{
			name:          "Fail - ใส่จำนวนสินค้าติดลบหรือเป็น 0",
			inputQuantity: 0,
			setupMockRepo: func() *mockCartRepo {
				// กรณีนี้ โค้ดควรจะเด้ง Error ทันทีโดยที่ไม่ทันได้เรียก DB เลย
				return &mockCartRepo{}
			},
			expectedError: "จำนวนสินค้าต้องมากกว่า 0",
		},
		{
			name:          "Success - มีตะกร้าอยู่แล้ว หยิบของใส่สำเร็จ",
			inputQuantity: 2,
			setupMockRepo: func() *mockCartRepo {
				return &mockCartRepo{
					// จำลองว่า DB หาตะกร้าเจอ (ID ตะกร้า = 99)
					mockGetCart: &domain.Cart{ID: 99, UserID: 1},
					mockAddErr:  nil, // บันทึกลงตะกร้าผ่านฉลุย
				}
			},
			expectedError: "",
		},
		{
			name:          "Success - ยังไม่มีตะกร้า ระบบสร้างให้ใหม่ แล้วหยิบใส่สำเร็จ",
			inputQuantity: 1,
			setupMockRepo: func() *mockCartRepo {
				return &mockCartRepo{
					mockGetCart: nil, // 👈 จำลองว่าหาตะกร้าไม่เจอ
					mockGetErr:  nil,
					mockCreate:  &domain.Cart{ID: 100, UserID: 1}, // 👈 DB เลยสร้างใบใหม่ให้ (ID = 100)
					mockAddErr:  nil,
				}
			},
			expectedError: "",
		},
		{
			name:          "Fail - Database พังตอนกำลังหยิบของใส่",
			inputQuantity: 1,
			setupMockRepo: func() *mockCartRepo {
				return &mockCartRepo{
					mockGetCart: &domain.Cart{ID: 99, UserID: 1},
					// จำลองว่า DB เน็ตหลุด หรือพังตอนกำลังรันคำสั่ง AddItem
					mockAddErr: errors.New("db connection lost"),
				}
			},
			expectedError: "db connection lost",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. โหลดตัวปลอมตามสถานการณ์
			mockRepo := tc.setupMockRepo()

			// 2. ประกอบร่าง Service
			cartSvc := usecase.NewCartUsecase(mockRepo)

			// 3. สั่งรันฟังก์ชัน
			err := cartSvc.AddToCart(context.Background(), 1, 55, tc.inputQuantity)

			// 4. ตรวจกระดาษคำตอบ
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
