package service

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"simple-clothes-shop/internal/domain"
)

type productService struct {
	repo         domain.ProductRepository
	categoryRepo domain.CategoryRepository
}

func NewProductService(repo domain.ProductRepository, catRepo domain.CategoryRepository) domain.ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: catRepo,
	}
}

func (s *productService) UpdateProduct(ctx context.Context, id uint, product *domain.Product) error {
	existingProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("ไม่พบสินค้าที่ต้องการแก้ไข")
	}

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
		_, err := s.categoryRepo.GetByID(ctx, product.CategoryID) // 👈 ส่ง ctx ต่อให้ categoryRepo
		if err != nil {
			return errors.New("ไม่พบหมวดหมู่สินค้าที่ระบุ")
		}
		existingProduct.CategoryID = product.CategoryID
	}
	if len(product.Images) > 0 {
		existingProduct.Images = product.Images
	}

	if len(product.Variants) > 0 {
		existingProduct.Variants = product.Variants
	}

	return s.repo.Update(ctx, id, existingProduct)
}

func (s *productService) CreateProduct(ctx context.Context, product *domain.Product) error {
	if product.Price <= 0 {
		return errors.New("ราคาพื้นฐานของสินค้าต้องมากกว่า 0 บาท")
	}
	if product.Stock < 0 {
		return errors.New("สต็อกสินค้าไม่สามารถติดลบได้")
	}
	for _, v := range product.Variants {
		if v.Price <= 0 {
			return errors.New("ราคาของสินค้าแต่ละสี/ไซส์ (Variant) ต้องมากกว่า 0 บาท")
		}
	}
	return s.repo.Create(ctx, product)
}

func (s *productService) RemoveProduct(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *productService) FetchAll(ctx context.Context) ([]domain.Product, error) {
	products, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for i := range products {
		cat, _ := s.categoryRepo.GetByID(ctx, products[i].CategoryID) // 👈
		products[i].Category = cat
	}
	return products, nil
}

func (s *productService) FetchByID(ctx context.Context, id uint) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cat, _ := s.categoryRepo.GetByID(ctx, product.CategoryID) // 👈
	product.Category = cat
	return product, nil
}

func (s *productService) FetchByCategoryID(ctx context.Context, categoryID uint) ([]domain.Product, error) {
	products, err := s.repo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	cat, _ := s.categoryRepo.GetByID(ctx, categoryID) // 👈
	for i := range products {
		products[i].Category = cat
	}
	return products, nil
}

func (s *productService) FetchWithFilter(ctx context.Context, categoryID *uint, minPrice *float64, maxPrice *float64, page int, limit int) ([]domain.Product, error) {
	offset := (page - 1) * limit
	products, err := s.repo.GetWithFilter(ctx, categoryID, minPrice, maxPrice, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range products {
		cat, _ := s.categoryRepo.GetByID(ctx, products[i].CategoryID) // 👈
		products[i].Category = cat
	}
	return products, nil
}

func (s *productService) RemoveVariant(ctx context.Context, variantID uint) error {
	err := s.repo.DeleteVariant(ctx, variantID)
	if err != nil {
		return errors.New("ไม่พบ Variant นี้ในระบบ (ลบไม่สำเร็จ)")
	}
	return nil
}
