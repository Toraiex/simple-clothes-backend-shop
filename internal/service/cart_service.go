package service

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"simple-clothes-shop/internal/domain"
)

type cartService struct {
	cartRepo domain.CartRepository
}

var _ domain.CartService = (*cartService)(nil)

func NewCartService(cartRepo domain.CartRepository) domain.CartService {
	return &cartService{
		cartRepo: cartRepo,
	}
}

func (s *cartService) GetMyCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID) // 👈 ส่ง ctx ต่อ
	if err != nil {
		return nil, errors.New("เกิดข้อผิดพลาดในการค้นหาตะกร้าสินค้า")
	}

	if cart == nil {
		cart, err = s.cartRepo.CreateCart(ctx, userID) // 👈 ส่ง ctx ต่อ
		if err != nil {
			return nil, errors.New("ไม่สามารถสร้างตะกร้าสินค้าใหม่ได้")
		}
	}

	items, err := s.cartRepo.GetCartItemsWithDetails(ctx, cart.ID) // 👈 ส่ง ctx ต่อ
	if err != nil {
		return nil, errors.New("เกิดข้อผิดพลาดในการดึงรายการสินค้า")
	}

	cart.Items = items
	return cart, nil
}

func (s *cartService) AddToCart(ctx context.Context, userID uint, variantID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("จำนวนสินค้าต้องมากกว่า 0")
	}

	cart, err := s.GetMyCart(ctx, userID) // 👈 ส่ง ctx ต่อ
	if err != nil {
		return err
	}

	return s.cartRepo.AddItem(ctx, cart.ID, variantID, quantity) // 👈 ส่ง ctx ต่อ
}

func (s *cartService) UpdateQuantity(ctx context.Context, userID uint, cartItemID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("จำนวนสินค้าต้องมากกว่า 0 (หากต้องการนำออก ให้กดปุ่มลบสินค้า)")
	}

	return s.cartRepo.UpdateItemQuantity(ctx, cartItemID, quantity) // 👈 ส่ง ctx ต่อ
}

func (s *cartService) RemoveFromCart(ctx context.Context, userID uint, cartItemID uint) error {
	return s.cartRepo.RemoveItem(ctx, cartItemID) // 👈 ส่ง ctx ต่อ
}
