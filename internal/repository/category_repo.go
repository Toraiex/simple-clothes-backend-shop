package repository

import (
	"time"

	"simple-clothes-shop/internal/domain"

	"gorm.io/gorm"
)

type CategoryModel struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"unique;not null"`
	Products  []ProductModel `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CategoryModel) TableName() string {
	return "categories"
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var models []CategoryModel

	if err := r.db.
		Preload("Products").
		Find(&models).Error; err != nil {
		return nil, err
	}

	var categories []domain.Category
	for _, m := range models {
		categories = append(categories, toDomainWithProducts(m))
	}

	return categories, nil
}

func (r *categoryRepository) GetByID(id uint) (*domain.Category, error) {
	var model CategoryModel
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}

	category := toDomain(model)
	return &category, nil
}

func (r *categoryRepository) Create(category *domain.Category) error {
	model := toModel(*category)

	if err := r.db.Create(&model).Error; err != nil {
		return err
	}

	category.ID = model.ID
	return nil
}

func (r *categoryRepository) Update(category *domain.Category) error {
	model := toModel(*category)
	return r.db.Save(&model).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&CategoryModel{}, id).Error
}

func toDomain(model CategoryModel) domain.Category {
	return domain.Category{
		ID:        model.ID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toModel(entity domain.Category) CategoryModel {
	return CategoryModel{
		ID:        entity.ID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
func toDomainWithProducts(model CategoryModel) domain.Category {
	var products []domain.Product

	for _, p := range model.Products {
		products = append(products, domain.Product{
			ID:    p.ID,
			Name:  p.Name,
			Price: p.Price,
		})
	}

	return domain.Category{
		ID:        model.ID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
func (r *productRepository) GetByCategoryID(categoryID uint) ([]domain.Product, error) {
	var models []ProductModel

	if err := r.db.
		Where("category_id = ?", categoryID).
		Preload("Variants").
		Find(&models).Error; err != nil {
		return nil, err
	}

	var products []domain.Product
	for _, m := range models {
		products = append(products, toDomainProduct(m))
	}

	return products, nil
}
