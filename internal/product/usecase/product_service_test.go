package usecase_test

import (
	"context"
	"testing"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/product/usecase"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// 1. สร้างตัวปลอม (Mock) สำหรับ Product (ใช้ร่วมกับ Category ของเดิมได้)
// =====================================================================
type mockProductRepoForTest struct {
	mockCreateErr error
}

func (m *mockProductRepoForTest) Create(ctx context.Context, product *domain.Product) error {
	return m.mockCreateErr
}
func (m *mockProductRepoForTest) GetAll(ctx context.Context) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepoForTest) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepoForTest) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepoForTest) GetWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, limit int, offset int) ([]domain.Product, error) {
	return nil, nil
}
func (m *mockProductRepoForTest) Update(ctx context.Context, id uint, product *domain.Product) error {
	return nil
}
func (m *mockProductRepoForTest) Delete(ctx context.Context, id uint) error        { return nil }
func (m *mockProductRepoForTest) DeleteVariant(ctx context.Context, id uint) error { return nil }

type mockCategoryRepoForProd struct{}

func (m *mockCategoryRepoForProd) GetAll(ctx context.Context) ([]domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepoForProd) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepoForProd) Create(ctx context.Context, category *domain.Category) error {
	return nil
}
func (m *mockCategoryRepoForProd) Update(ctx context.Context, category *domain.Category) error {
	return nil
}
func (m *mockCategoryRepoForProd) Delete(ctx context.Context, id uint) error { return nil }

// =====================================================================
// 2. เขียนเคสทดสอบสำหรับ CreateProduct
// =====================================================================
func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name          string
		inputProduct  *domain.Product
		expectedError string
	}{
		{
			name: "Fail - ราคาติดลบหรือเท่ากับ 0",
			inputProduct: &domain.Product{
				Name:  "กางเกงยีนส์",
				Price: 0, // 👈 ผิดกฎ
				Stock: 10,
			},
			expectedError: "ราคาพื้นฐานของสินค้าต้องมากกว่า 0 บาท",
		},
		{
			name: "Fail - สต็อกสินค้าติดลบ",
			inputProduct: &domain.Product{
				Name:  "กางเกงยีนส์",
				Price: 500,
				Stock: -5, // 👈 ผิดกฎ
			},
			expectedError: "สต็อกสินค้าไม่สามารถติดลบได้",
		},
		{
			name: "Fail - ราคา Variant (ตัวเลือกสี/ไซส์) ติดลบหรือเป็น 0",
			inputProduct: &domain.Product{
				Name:  "กางเกงยีนส์",
				Price: 500,
				Stock: 10,
				Variants: []domain.ProductVariant{
					{SKU: "JEAN-B-M", Price: 0, Stock: 5}, // 👈 ผิดกฎ
				},
			},
			expectedError: "ราคาของสินค้าแต่ละสี/ไซส์ (Variant) ต้องมากกว่า 0 บาท",
		},
		{
			name: "Success - สินค้าถูกต้องตามกฎทุกอย่าง",
			inputProduct: &domain.Product{
				Name:  "กางเกงยีนส์",
				Price: 500,
				Stock: 10,
				Variants: []domain.ProductVariant{
					{SKU: "JEAN-B-M", Price: 500, Stock: 5},
				},
			},
			expectedError: "", // ไม่ควรมี Error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockProd := &mockProductRepoForTest{}
			mockCat := &mockCategoryRepoForProd{}

			prodSvc := usecase.NewProductUsecase(mockProd, mockCat)

			err := prodSvc.CreateProduct(context.Background(), tc.inputProduct)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
		})
	}
}
