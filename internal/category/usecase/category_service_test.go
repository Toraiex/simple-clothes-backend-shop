package usecase_test // ใช้ _test ต่อท้ายเพื่อแยกบริบทของการเทสออกมา

import (
	"context"
	"errors"
	"testing"

	"simple-clothes-shop/internal/category/usecase"
	"simple-clothes-shop/internal/domain"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// 1. สร้างตัวปลอม (Mock) สำหรับ Repository
// เราไม่ต้องการต่อ Database จริงๆ เลยสร้าง Struct ปลอมๆ ขึ้นมาสวมรอย Interface
// =====================================================================

type mockCategoryRepo struct {
	// ตัวแปรนี้เอาไว้สั่งให้ตัวปลอม Return ค่า Error ตามที่เราต้องการจำลอง
	mockCreateError error
}

// สร้างฟังก์ชันปลอมให้ครบตามที่สัญญา (Interface) ระบุไว้
func (m *mockCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	// จำลองการเซฟลง DB ถ้ามี error ที่ตั้งไว้ก็ให้ return ออกไปเลย
	return m.mockCreateError
}
func (m *mockCategoryRepo) GetAll(ctx context.Context) ([]domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) Update(ctx context.Context, category *domain.Category) error { return nil }
func (m *mockCategoryRepo) Delete(ctx context.Context, id uint) error                   { return nil }

// สร้างตัวปลอมของ ProductRepo ด้วย (เพราะ CategoryUsecase ต้องการใช้)
type mockProductRepo struct{}

func (m *mockProductRepo) GetAll(ctx context.Context) ([]domain.Product, error) { return nil, nil }
func (m *mockProductRepo) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) GetWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, limit int, offset int) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepo) Create(ctx context.Context, product *domain.Product) error { return nil }
func (m *mockProductRepo) Update(ctx context.Context, id uint, product *domain.Product) error {
	return nil
}
func (m *mockProductRepo) Delete(ctx context.Context, id uint) error        { return nil }
func (m *mockProductRepo) DeleteVariant(ctx context.Context, id uint) error { return nil }

// =====================================================================
// 2. เขียนเคสทดสอบ (Test Cases)
// =====================================================================

func TestCreateCategory(t *testing.T) {
	// เราจะใช้ท่า Table-Driven Test (สร้างตารางจำลองสถานการณ์ต่างๆ)
	tests := []struct {
		name          string // ชื่อเคสทดสอบ
		inputName     string // ชื่อหมวดหมู่ที่จะส่งเข้าไป
		mockRepoError error  // อยากให้ DB ปลอมจำลองพังไหม?
		expectedError string // ข้อความ Error ที่คาดหวังว่าจะได้รับกลับมา ("" คือสำเร็จ)
	}{
		{
			name:          "Success - สร้างหมวดหมู่สำเร็จ",
			inputName:     "Sneakers",
			mockRepoError: nil,
			expectedError: "",
		},
		{
			name:          "Fail - ชื่อหมวดหมู่เป็นค่าว่างเปล่า",
			inputName:     "   ", // ส่ง space bar ล้วนๆ ไป
			mockRepoError: nil,
			expectedError: "ชื่อหมวดหมู่ห้ามเป็นค่าว่าง",
		},
		{
			name:          "Fail - ชื่อหมวดหมู่ซ้ำในระบบ",
			inputName:     "Clothes",
			mockRepoError: errors.New("pq: duplicate key value violates unique constraint"), // จำลอง DB แดง
			expectedError: "ชื่อหมวดหมู่นี้มีอยู่ในระบบแล้ว",
		},
	}

	// ลูปทดสอบทีละสถานการณ์
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. จัดเตรียมตัวปลอม
			catRepo := &mockCategoryRepo{mockCreateError: tc.mockRepoError}
			prodRepo := &mockProductRepo{}

			// 2. สร้าง Service ของจริง (แต่ฉีด DB ปลอมเข้าไปแทน)
			categorySvc := usecase.NewCategoryUsecase(catRepo, prodRepo)

			// 3. สั่งรันฟังก์ชันที่ต้องการทดสอบ!
			err := categorySvc.CreateCategory(context.Background(), tc.inputName)

			// 4. ตรวจข้อสอบ (Assert)
			if tc.expectedError == "" {
				// กรณีคาดหวังว่าต้องสำเร็จ -> err ต้องเป็น nil
				assert.NoError(t, err)
			} else {
				// กรณีคาดหวังว่าต้องพัง -> err ต้องไม่เป็น nil และข้อความต้องตรงเป๊ะ
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
