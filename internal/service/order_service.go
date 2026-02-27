package service

import (
	"context" // 👈 เพิ่ม context
	"errors"
	"simple-clothes-shop/internal/domain"
)

type orderService struct {
	repo     domain.OrderRepository
	cartRepo domain.CartRepository
}

func NewOrderService(repo domain.OrderRepository, cartRepo domain.CartRepository) domain.OrderService {
	return &orderService{
		repo:     repo,
		cartRepo: cartRepo,
	}
}

func (s *orderService) Checkout(ctx context.Context, userID uint) error {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID) // 👈 ส่ง ctx ต่อ
	if err != nil || cart == nil {
		return errors.New("ไม่พบตะกร้าสินค้า")
	}

	items, err := s.cartRepo.GetCartItemsWithDetails(ctx, cart.ID) // 👈 ส่ง ctx ต่อ
	if err != nil {
		return errors.New("เกิดข้อผิดพลาดในการดึงรายการสินค้า")
	}

	if len(items) == 0 {
		return errors.New("ตะกร้าสินค้าว่างเปล่า ไม่สามารถสั่งซื้อได้")
	}

	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Variant.Price * float64(item.Quantity)
	}

	_, err = s.repo.CreateOrderFromCart(ctx, userID, items, totalAmount) // 👈 ส่ง ctx ต่อ
	return err
}

func (s *orderService) GetByUserID(ctx context.Context, userID uint) ([]domain.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *orderService) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) GetByIDForUser(ctx context.Context, userID uint, orderID uint) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, errors.New("คุณไม่มีสิทธิ์ดูคำสั่งซื้อนี้")
	}
	return order, nil
}

func (s *orderService) CancelOrder(ctx context.Context, userID uint, orderID uint) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.UserID != userID {
		return errors.New("คุณไม่มีสิทธิ์ยกเลิกคำสั่งซื้อนี้")
	}
	if order.Status != domain.StatusPending {
		return errors.New("สามารถยกเลิกได้เฉพาะคำสั่งซื้อที่อยู่ในสถานะ pending เท่านั้น")
	}
	return s.repo.CancelAndRestoreStock(ctx, orderID)
}

func (s *orderService) AdminUpdateStatus(ctx context.Context, orderID uint, newStatus domain.OrderStatus) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status == domain.StatusCanceled {
		return errors.New("ไม่สามารถแก้ไขคำสั่งซื้อที่ถูกยกเลิกแล้ว")
	}
	if order.Status == domain.StatusShipped {
		return errors.New("ไม่สามารถแก้ไขคำสั่งซื้อที่จัดส่งแล้ว")
	}

	switch order.Status {
	case domain.StatusPending:
		if newStatus != domain.StatusPaid && newStatus != domain.StatusCanceled {
			return errors.New("สถานะไม่ถูกต้องสำหรับคำสั่งซื้อที่เป็น pending")
		}
	case domain.StatusPaid:
		if newStatus != domain.StatusShipped {
			return errors.New("คำสั่งซื้อที่ชำระเงินแล้วสามารถเปลี่ยนเป็น shipped เท่านั้น")
		}
	}
	return s.repo.UpdateStatus(ctx, orderID, newStatus)
}
