package service

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"strings"

	"simple-clothes-shop/internal/domain"
)

type categoryService struct {
	repo        domain.CategoryRepository
	productRepo domain.ProductRepository
}

func NewCategoryService(repo domain.CategoryRepository, productRepo domain.ProductRepository) domain.CategoryService {
	return &categoryService{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *categoryService) FetchAll(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *categoryService) GetCategory(ctx context.Context, id uint) (*domain.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 👈 ส่ง ctx ต่อให้ productRepo ด้วย
	products, err := s.productRepo.GetByCategoryID(ctx, id)
	if err == nil {
		category.Products = products
	}

	return category, nil
}

func (s *categoryService) CreateCategory(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง")
	}

	category := &domain.Category{Name: name}
	err := s.repo.Create(ctx, category)

	if err != nil && strings.Contains(err.Error(), "unique constraint") {
		return errors.New("ชื่อหมวดหมู่นี้มีอยู่ในระบบแล้ว")
	}

	return err
}

func (s *categoryService) UpdateCategory(ctx context.Context, id uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง")
	}

	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("ไม่พบหมวดหมู่นี้ในระบบ")
	}

	category.Name = name
	err = s.repo.Update(ctx, category)

	if err != nil && strings.Contains(err.Error(), "unique constraint") {
		return errors.New("ไม่สามารถเปลี่ยนเป็นชื่อนี้ได้ เนื่องจากมีอยู่ในระบบแล้ว")
	}

	return err
}

func (s *categoryService) RemoveCategory(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
