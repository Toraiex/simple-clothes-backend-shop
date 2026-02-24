package service

import (
	"errors"
	"simple-clothes-shop/internal/domain"
)

type orderService struct {
	repo     domain.OrderRepository
	cartRepo domain.CartRepository // 👈 เปลี่ยนมาใช้ CartRepo แทน ProductRepo
}

func NewOrderService(repo domain.OrderRepository, cartRepo domain.CartRepository) domain.OrderService {
	return &orderService{
		repo:     repo,
		cartRepo: cartRepo,
	}
}

// =================================================================
// 🛒 1. ฟังก์ชัน Checkout (ดึงของจากตะกร้ามาคิดเงิน)
// =================================================================
func (s *orderService) Checkout(userID uint) error {
	// 1. หาตะกร้าของลูกค้าคนนี้
	cart, err := s.cartRepo.GetCartByUserID(userID)
	if err != nil || cart == nil {
		return errors.New("ไม่พบตะกร้าสินค้า")
	}

	// 2. กวาดของในตะกร้ามาดู
	items, err := s.cartRepo.GetCartItemsWithDetails(cart.ID)
	if err != nil {
		return errors.New("เกิดข้อผิดพลาดในการดึงรายการสินค้า")
	}

	if len(items) == 0 {
		return errors.New("ตะกร้าสินค้าว่างเปล่า ไม่สามารถสั่งซื้อได้")
	}

	// 3. คำนวณยอดรวมทั้งหมด (Total Amount)
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Variant.Price * float64(item.Quantity)
	}

	// 4. ส่งไม้ต่อให้ Repo จัดการตัดสต็อก สร้างบิล และล้างตะกร้าแบบ 4-in-1!
	_, err = s.repo.CreateOrderFromCart(userID, items, totalAmount)
	return err
}

// =================================================================
// ฟังก์ชันอื่นๆ (เหมือนเดิมเป๊ะ แค่จัดให้เป็นระเบียบ)
// =================================================================

func (s *orderService) GetByUserID(userID uint) ([]domain.Order, error) {
	return s.repo.GetByUserID(userID)
}

func (s *orderService) GetByID(id uint) (*domain.Order, error) {
	return s.repo.GetByID(id)
}

func (s *orderService) GetByIDForUser(userID uint, orderID uint) (*domain.Order, error) {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, errors.New("คุณไม่มีสิทธิ์ดูคำสั่งซื้อนี้")
	}
	return order, nil
}

func (s *orderService) CancelOrder(userID uint, orderID uint) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order.UserID != userID {
		return errors.New("คุณไม่มีสิทธิ์ยกเลิกคำสั่งซื้อนี้")
	}
	if order.Status != domain.StatusPending {
		return errors.New("สามารถยกเลิกได้เฉพาะคำสั่งซื้อที่อยู่ในสถานะ pending เท่านั้น")
	}
	return s.repo.CancelAndRestoreStock(orderID)
}

func (s *orderService) AdminUpdateStatus(orderID uint, newStatus domain.OrderStatus) error {
	order, err := s.repo.GetByID(orderID)
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
	return s.repo.UpdateStatus(orderID, newStatus)
}
