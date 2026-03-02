package postgres // ✅ ใหม่: เปลี่ยนชื่อแพ็กเกจให้ตรงกับโฟลเดอร์

import (
	"context"
	"errors"
	"simple-clothes-shop/internal/domain"
	"strings"
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
	return s.repo.GetAll(ctx)
}

func (s *categoryUsecase) GetCategory(ctx context.Context, id uint) (*domain.Category, error) {
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

func (s *categoryUsecase) CreateCategory(ctx context.Context, name string) error {
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

func (s *categoryUsecase) UpdateCategory(ctx context.Context, id uint, name string) error {
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

func (s *categoryUsecase) RemoveCategory(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
