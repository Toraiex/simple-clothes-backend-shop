package usecase

import (
	"context"
	"errors"
	"fmt"
	"simple-clothes-shop/internal/domain"
)

type cartUsecase struct {
	cartRepo domain.CartRepository
}

var _ domain.CartUsecase = (*cartUsecase)(nil)

func NewCartUsecase(cartRepo domain.CartRepository) domain.CartUsecase {
	return &cartUsecase{
		cartRepo: cartRepo,
	}
}

func (s *cartUsecase) GetMyCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternalServerError // 💡 ปิดรอยรั่ว DB
	}

	if cart == nil {
		cart, err = s.cartRepo.CreateCart(ctx, userID)
		if err != nil {
			return nil, domain.ErrInternalServerError
		}
	}

	items, err := s.cartRepo.GetCartItemsWithDetails(ctx, cart.ID)
	if err != nil {
		return nil, domain.ErrInternalServerError
	}

	cart.Items = items
	return cart, nil
}

func (s *cartUsecase) AddToCart(ctx context.Context, userID uint, variantID uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("จำนวนสินค้าต้องมากกว่า 0: %w", domain.ErrBadParamInput) // 💡 ห่อ Error ลูกค้าพิมพ์ผิด
	}

	cart, err := s.GetMyCart(ctx, userID)
	if err != nil {
		return err // ส่งต่อ Error จาก GetMyCart ได้เลย
	}

	err = s.cartRepo.AddItem(ctx, cart.ID, variantID, quantity)
	if err != nil {
		return err
	}
	return nil
}

func (s *cartUsecase) UpdateQuantity(ctx context.Context, userID uint, cartItemID uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("จำนวนสินค้าต้องมากกว่า 0 (หากต้องการนำออก ให้กดปุ่มลบสินค้า): %w", domain.ErrBadParamInput)
	}

	err := s.cartRepo.UpdateItemQuantity(ctx, cartItemID, quantity)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบรายการสินค้านี้ในตะกร้า: %w", domain.ErrNotFound)
		}
		return domain.ErrInternalServerError
	}
	return nil
}

func (s *cartUsecase) RemoveFromCart(ctx context.Context, userID uint, cartItemID uint) error {
	err := s.cartRepo.RemoveItem(ctx, cartItemID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("ไม่พบรายการสินค้านี้ในตะกร้า: %w", domain.ErrNotFound)
		}
		return err
	}
	return nil
}
