package service

import (
	"errors"
	"simple-clothes-shop/internal/domain"
	"sort"
)

type orderService struct {
	repo        domain.OrderRepository
	productRepo domain.ProductRepository
}

func (s *orderService) CreateOrder(userID uint, items []domain.OrderItem) error {
	var total float64
	if len(items) == 0 {
		return errors.New("order ต้องมีอย่างน้อย 1 สินค้า")
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})

	for i, item := range items {

		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return err
		}

		items[i].UnitPrice = product.Price
		items[i].Total = product.Price * float64(item.Quantity)

		total += items[i].Total
	}

	order := &domain.Order{
		UserID: userID,
		Status: domain.StatusPending,
		Items:  items,
		Total:  total,
	}

	return s.repo.Create(order)
}
func NewOrderService(
	repo domain.OrderRepository,
	productRepo domain.ProductRepository,
) domain.OrderService {
	return &orderService{
		repo:        repo,
		productRepo: productRepo,
	}
}
func (s *orderService) GetByUserID(userID uint) ([]domain.Order, error) {
	return s.repo.GetByUserID(userID)
}

func (s *orderService) UpdateStatus(id uint, status domain.OrderStatus) error {

	order, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// กฎธุรกิจตัวอย่าง
	if order.Status == domain.StatusShipped {
		return errors.New("ไม่สามารถเปลี่ยนสถานะหลังจากจัดส่งแล้ว")
	}

	if order.Status == domain.StatusCanceled {
		return errors.New("ไม่สามารถเปลี่ยนสถานะหลังจากยกเลิกแล้ว")
	}

	return s.repo.UpdateStatus(id, status)
}
func (s *orderService) CancelOrder(userID uint, orderID uint) error {

	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return err
	}

	// 🔐 เช็คว่าเป็นเจ้าของ order ไหม
	if order.UserID != userID {
		return errors.New("คุณไม่มีสิทธิ์ยกเลิกคำสั่งซื้อนี้")
	}

	// 🔒 ยกเลิกได้เฉพาะ pending
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

	// ❌ ห้ามแก้ถ้า canceled
	if order.Status == domain.StatusCanceled {
		return errors.New("ไม่สามารถแก้ไขคำสั่งซื้อที่ถูกยกเลิกแล้ว")
	}

	// ❌ shipped แล้วห้ามแก้
	if order.Status == domain.StatusShipped {
		return errors.New("ไม่สามารถแก้ไขคำสั่งซื้อที่จัดส่งแล้ว")
	}

	// Workflow ที่อนุญาต
	switch order.Status {

	case domain.StatusPending:
		if newStatus != domain.StatusPaid &&
			newStatus != domain.StatusCanceled {
			return errors.New("สถานะไม่ถูกต้องสำหรับคำสั่งซื้อที่เป็น pending")
		}

	case domain.StatusPaid:
		if newStatus != domain.StatusShipped {
			return errors.New("คำสั่งซื้อที่ชำระเงินแล้วสามารถเปลี่ยนเป็น shipped เท่านั้น")
		}
	}

	return s.repo.UpdateStatus(orderID, newStatus)
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

func (s *orderService) GetByID(id uint) (*domain.Order, error) {
	return s.repo.GetByID(id)
}
