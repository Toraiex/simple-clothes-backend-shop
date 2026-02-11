package service

import "simple-clothes-shop/internal/domain"

type categoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) FetchAll() ([]domain.Category, error) {
	return s.repo.GetAll()
}

// ✅ ฟังก์ชันนี้ต้องอยู่ที่นี่!
func (s *categoryService) GetCategory(id uint) (*domain.Category, error) {
	return s.repo.GetByID(id)
}

func (s *categoryService) CreateCategory(name string) error {
	category := &domain.Category{Name: name}
	return s.repo.Create(category)
}

func (s *categoryService) RemoveCategory(id uint) error {
	return s.repo.Delete(id)
}

func (s *categoryService) UpdateCategory(id uint, name string) error {
	category, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	category.Name = name

	return s.repo.Update(id, category)
}
