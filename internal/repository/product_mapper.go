package repository

import "simple-clothes-shop/internal/domain"

func toDomainProduct(m ProductModel) domain.Product {
	var variants []domain.ProductVariant

	for _, v := range m.Variants {
		variants = append(variants, domain.ProductVariant{
			ID:        v.ID,
			ProductID: v.ProductID,
			Color:     v.Color,
			Size:      v.Size,
			Price:     v.Price,
			Stock:     v.Stock,
		})
	}

	return domain.Product{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Stock:       m.Stock,
		CategoryID:  m.CategoryID,
		Image:       m.Image,
		Variants:    variants,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
