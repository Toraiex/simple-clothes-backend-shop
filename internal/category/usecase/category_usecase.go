package usecase

import (
	"context"
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"
	"strings"

	"github.com/sirupsen/logrus"
)

type categoryUsecase struct {
	repo        domain.CategoryRepository
	productRepo domain.ProductRepository
}

func NewCategoryUsecase(repo domain.CategoryRepository, productRepo domain.ProductRepository) domain.CategoryUsecase {
	return &categoryUsecase{
		repo:        repo,
		productRepo: productRepo,
	}
}

func (s *categoryUsecase) FetchAll(ctx context.Context) ([]domain.Category, error) {
	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}
	return categories, nil
}

func (s *categoryUsecase) GetCategory(ctx context.Context, id uint) (*domain.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("ไม่พบหมวดหมู่นี้ในระบบ: %w", domain.ErrNotFound)
		}
		return nil, domain.ErrInternalServerError
	}

	products, err := s.productRepo.GetByCategoryID(ctx, id)
	if err == nil {
		category.Products = products
	}

	return category, nil
}

func (s *categoryUsecase) CreateCategory(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง: %w", domain.ErrBadParamInput)
	}

	category := &domain.Category{Name: name}
	err := s.repo.Create(ctx, category)

	if err != nil {
		// 💡 ดักจับฐานข้อมูลบ่นเรื่องชื่อซ้ำ
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("ชื่อหมวดหมู่นี้มีอยู่ในระบบแล้ว: %w", domain.ErrConflict)
		}
		logrus.Error(err) // ถ้าเป็น Error อื่น ค่อยปริ้น Log แดง
		return domain.ErrInternalServerError
	}

	return nil
}

func (s *categoryUsecase) UpdateCategory(ctx context.Context, id uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง: %w", domain.ErrBadParamInput)
	}

	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบหมวดหมู่นี้ในระบบ: %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError
	}

	category.Name = name
	err = s.repo.Update(ctx, category)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("ไม่สามารถเปลี่ยนเป็นชื่อนี้ได้ เนื่องจากมีอยู่ในระบบแล้ว: %w", domain.ErrConflict)
		}
		logrus.Error(err)
		return domain.ErrInternalServerError
	}

	return nil
}

func (s *categoryUsecase) RemoveCategory(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบข้อมูลหมวดหมู่นี้ในระบบ (ลบไม่สำเร็จ): %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError
	}
	return nil
}
