package service_test

import (
	"context"
	"errors"
	"testing"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/service"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// 1. สร้างตัวปลอม (Mock) สำหรับ Order และ Cart
// =====================================================================

// --- ตัวปลอมของ Order Repo ---
type mockOrderRepo struct {
	mockCreateOrder    *domain.Order
	mockCreateOrderErr error
}

func (m *mockOrderRepo) CreateOrderFromCart(ctx context.Context, userID uint, cartItems []domain.CartItem, totalAmount float64) (*domain.Order, error) {
	return m.mockCreateOrder, m.mockCreateOrderErr
}
func (m *mockOrderRepo) GetByID(ctx context.Context, id uint) (*domain.Order, error) { return nil, nil }
func (m *mockOrderRepo) GetByUserID(ctx context.Context, userID uint) ([]domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uint, status domain.OrderStatus) error {
	return nil
}
func (m *mockOrderRepo) CancelAndRestoreStock(ctx context.Context, orderID uint) error { return nil }

// --- ตัวปลอมของ Cart Repo (ทำแยกมาเพื่อใช้กับเทส Order โดยเฉพาะ) ---
type mockCartRepoForOrder struct {
	mockGetCart     *domain.Cart
	mockGetCartErr  error
	mockGetItems    []domain.CartItem
	mockGetItemsErr error
}

func (m *mockCartRepoForOrder) GetCartByUserID(ctx context.Context, userID uint) (*domain.Cart, error) {
	return m.mockGetCart, m.mockGetCartErr
}
func (m *mockCartRepoForOrder) GetCartItemsWithDetails(ctx context.Context, cartID uint) ([]domain.CartItem, error) {
	return m.mockGetItems, m.mockGetItemsErr
}
func (m *mockCartRepoForOrder) CreateCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	return nil, nil
}
func (m *mockCartRepoForOrder) AddItem(ctx context.Context, cartID uint, variantID uint, quantity int) error {
	return nil
}
func (m *mockCartRepoForOrder) UpdateItemQuantity(ctx context.Context, cartItemID uint, quantity int) error {
	return nil
}
func (m *mockCartRepoForOrder) RemoveItem(ctx context.Context, cartItemID uint) error { return nil }
func (m *mockCartRepoForOrder) ClearCart(ctx context.Context, cartID uint) error      { return nil }

// =====================================================================
// 2. เขียนเคสทดสอบสำหรับฟังก์ชัน Checkout
// =====================================================================

func TestCheckout(t *testing.T) {
	// สร้างตัวแปรสมมติเอาไว้ใช้ในเทส
	testUserID := uint(1)
	testCart := &domain.Cart{ID: 10, UserID: testUserID}

	// จำลองสินค้าในตะกร้า 2 ชิ้น (ตัวละ 500 บาท จำนวน 2 ชิ้น = รวม 1000 บาท)
	testCartItems := []domain.CartItem{
		{
			VariantID: 99,
			Quantity:  2,
			Variant:   &domain.ProductVariant{Price: 500.00},
		},
	}

	tests := []struct {
		name           string
		setupCartMock  func() *mockCartRepoForOrder // ตั้งค่าพฤติกรรมตะกร้าปลอม
		setupOrderMock func() *mockOrderRepo        // ตั้งค่าพฤติกรรมระบบสั่งซื้อปลอม
		expectedError  string
	}{
		{
			name: "Fail - ไม่พบตะกร้าสินค้า (หาใน DB ไม่เจอ)",
			setupCartMock: func() *mockCartRepoForOrder {
				return &mockCartRepoForOrder{
					mockGetCart: nil, // คืนค่าตะกร้าเป็นค่าว่าง
				}
			},
			setupOrderMock: func() *mockOrderRepo { return &mockOrderRepo{} },
			expectedError:  "ไม่พบตะกร้าสินค้า",
		},
		{
			name: "Fail - ตะกร้าว่างเปล่า ไม่มีของอยู่เลย",
			setupCartMock: func() *mockCartRepoForOrder {
				return &mockCartRepoForOrder{
					mockGetCart:  testCart,
					mockGetItems: []domain.CartItem{}, // คืนค่ารายการสินค้าเป็น Array ว่าง
				}
			},
			setupOrderMock: func() *mockOrderRepo { return &mockOrderRepo{} },
			expectedError:  "ตะกร้าสินค้าว่างเปล่า ไม่สามารถสั่งซื้อได้",
		},
		{
			name: "Fail - สต็อกสินค้าไม่พอ (Database แจ้งเตือนกลับมาตอนกำลังสร้าง Order)",
			setupCartMock: func() *mockCartRepoForOrder {
				return &mockCartRepoForOrder{
					mockGetCart:  testCart,
					mockGetItems: testCartItems,
				}
			},
			setupOrderMock: func() *mockOrderRepo {
				return &mockOrderRepo{
					// จำลองว่า DB ตัดสต็อกไม่สำเร็จ แล้วโยน Error ออกมา
					mockCreateOrderErr: errors.New("สินค้า 'เสื้อ' สี ดำ ไซส์ M ในสต็อกมีไม่เพียงพอ"),
				}
			},
			expectedError: "สินค้า 'เสื้อ' สี ดำ ไซส์ M ในสต็อกมีไม่เพียงพอ",
		},
		{
			name: "Success - Checkout สำเร็จ ตัดเงินผ่าน",
			setupCartMock: func() *mockCartRepoForOrder {
				return &mockCartRepoForOrder{
					mockGetCart:  testCart,
					mockGetItems: testCartItems,
				}
			},
			setupOrderMock: func() *mockOrderRepo {
				return &mockOrderRepo{
					mockCreateOrderErr: nil, // DB บันทึกสำเร็จ
				}
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. โหลดตัวปลอม
			mockCart := tc.setupCartMock()
			mockOrder := tc.setupOrderMock()

			// 2. ประกอบร่าง Order Service (ฉีด Repo ปลอม 2 ตัวเข้าไป)
			orderSvc := service.NewOrderService(mockOrder, mockCart)

			// 3. สั่งรันฟังก์ชัน
			err := orderSvc.Checkout(context.Background(), testUserID)

			// 4. ตรวจสอบกระดาษคำตอบ
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
