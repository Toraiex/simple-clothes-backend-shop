package service

import (
	"errors"
	"strings"

	"simple-clothes-shop/internal/domain"
)

type categoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) FetchAll() ([]domain.Category, error) {
	return s.repo.GetAll()
}

func (s *categoryService) GetCategory(id uint) (*domain.Category, error) {
	return s.repo.GetByID(id)
}

func (s *categoryService) CreateCategory(name string) error {
	// 1. ลบช่องว่างหน้า-หลังทิ้ง ป้องกันคนพิมพ์สเปซบาร์มาเฉยๆ ("   ")
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง")
	}

	category := &domain.Category{Name: name}
	err := s.repo.Create(category)

	// 2. จับ Error จาก Database กรณีชื่อซ้ำ
	if err != nil && strings.Contains(err.Error(), "unique constraint") {
		return errors.New("ชื่อหมวดหมู่นี้มีอยู่ในระบบแล้ว")
	}

	return err
}

func (s *categoryService) UpdateCategory(id uint, name string) error {
	// 1. ดักค่าว่างเหมือนตอน Create
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("ชื่อหมวดหมู่ห้ามเป็นค่าว่าง")
	}

	// 2. Fetch ของเก่ามาดู (โค้ดเดิมคุณทำไว้ดีแล้ว ป้องกันการแก้มั่ว)
	category, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("ไม่พบหมวดหมู่นี้ในระบบ")
	}

	category.Name = name
	err = s.repo.Update(category)

	// 3. จับ Error ชื่อซ้ำตอนอัปเดต
	if err != nil && strings.Contains(err.Error(), "unique constraint") {
		return errors.New("ไม่สามารถเปลี่ยนเป็นชื่อนี้ได้ เนื่องจากมีอยู่ในระบบแล้ว")
	}

	return err
}

func (s *categoryService) RemoveCategory(id uint) error {
	// เราใช้ท่า RowsAffected ใน Repo มาแล้ว ตรงนี้ส่งค่า err กลับไปให้ Handler จัดการต่อได้เลย
	return s.repo.Delete(id)
}
