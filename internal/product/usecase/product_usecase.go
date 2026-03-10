package usecase

import (
	"context"
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"
)

type productUsecase struct {
	repo         domain.ProductRepository
	categoryRepo domain.CategoryRepository
}

func NewProductUsecase(repo domain.ProductRepository, catRepo domain.CategoryRepository) domain.ProductUsecase {
	return &productUsecase{
		repo:         repo,
		categoryRepo: catRepo,
	}
}

func (s *productUsecase) UpdateProduct(ctx context.Context, id uint, product *domain.Product) error {
	existingProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบสินค้าที่ต้องการแก้ไข: %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError // 💡 ปิดรอยรั่ว DB
	}

	// ... (โค้ดอัปเดต Field ย่อยๆ ของคุณ เหมือนเดิมเป๊ะเลยครับ ขอละไว้เพื่อความสั้น) ...
	if product.Name != "" {
		existingProduct.Name = product.Name
	}
	if product.Description != "" {
		existingProduct.Description = product.Description
	}
	if product.Price > 0 {
		existingProduct.Price = product.Price
	}
	if product.Stock >= 0 {
		existingProduct.Stock = product.Stock
	}

	if product.CategoryID != 0 {
		_, err := s.categoryRepo.GetByID(ctx, product.CategoryID)
		if err != nil {
			return fmt.Errorf("ไม่พบหมวดหมู่สินค้าที่ระบุ: %w", domain.ErrBadParamInput) // 💡 ห่อ Error ลูกค้าพิมพ์หมวดหมู่ผิด
		}
		existingProduct.CategoryID = product.CategoryID
	}

	if len(product.Images) > 0 {
		existingProduct.Images = product.Images
	}
	if len(product.Variants) > 0 {
		existingProduct.Variants = product.Variants
	}

	err = s.repo.Update(ctx, id, existingProduct)
	if err != nil {
		return domain.ErrInternalServerError
	}
	return nil
}

func (s *productUsecase) CreateProduct(ctx context.Context, product *domain.Product) error {
	if product.Price <= 0 {
		return fmt.Errorf("ราคาพื้นฐานของสินค้าต้องมากกว่า 0 บาท: %w", domain.ErrBadParamInput)
	}
	if product.Stock < 0 {
		return fmt.Errorf("สต็อกสินค้าไม่สามารถติดลบได้: %w", domain.ErrBadParamInput)
	}
	for _, v := range product.Variants {
		if v.Price <= 0 {
			return fmt.Errorf("ราคาของสินค้าแต่ละสี/ไซส์ (Variant) ต้องมากกว่า 0 บาท: %w", domain.ErrBadParamInput)
		}
	}

	err := s.repo.Create(ctx, product)
	if err != nil {
		return err
	}
	return nil
}

func (s *productUsecase) RemoveProduct(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบสินค้านี้ในระบบ: %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError
	}
	return nil
}

func (s *productUsecase) FetchAll(ctx context.Context) ([]domain.Product, error) {
	products, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	for i := range products {
		cat, _ := s.categoryRepo.GetByID(ctx, products[i].CategoryID)
		products[i].Category = cat
	}
	return products, nil
}

func (s *productUsecase) FetchByID(ctx context.Context, id uint) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("ไม่พบสินค้าที่คุณต้องการ: %w", domain.ErrNotFound)
		}
		return nil, domain.ErrInternalServerError
	}

	cat, _ := s.categoryRepo.GetByID(ctx, product.CategoryID)
	product.Category = cat
	return product, nil
}

func (s *productUsecase) FetchByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	products, err := s.repo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	cat, _ := s.categoryRepo.GetByID(ctx, categoryID)
	for i := range products {
		products[i].Category = cat
	}
	return products, nil
}

func (s *productUsecase) FetchWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, page int, limit int) ([]domain.Product, error) {
	offset := (page - 1) * limit
	products, err := s.repo.GetWithFilter(ctx, categoryID, minPrice, maxPrice, limit, offset)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	for i := range products {
		cat, _ := s.categoryRepo.GetByID(ctx, products[i].CategoryID)
		products[i].Category = cat
	}
	return products, nil
}

func (s *productUsecase) RemoveVariant(ctx context.Context, variantID uint) error {
	err := s.repo.DeleteVariant(ctx, variantID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบ Variant นี้ในระบบ (ลบไม่สำเร็จ): %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError
	}
	return nil
}
