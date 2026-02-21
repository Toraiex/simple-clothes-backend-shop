package service

import (
	"errors"
	"simple-clothes-shop/internal/domain"
)

type cartService struct {
	cartRepo domain.CartRepository
}

// เช็คให้ชัวร์ว่า implement ครบตามสัญญา
var _ domain.CartService = (*cartService)(nil)

func NewCartService(cartRepo domain.CartRepository) domain.CartService {
	return &cartService{
		cartRepo: cartRepo,
	}
}

// -----------------------------------------------------------------
// 1. ดึงข้อมูลตะกร้าของฉัน (และสร้างให้ถ้ายังไม่มี)
// -----------------------------------------------------------------
func (s *cartService) GetMyCart(userID uint) (*domain.Cart, error) {
	// 1. ลองหาตะกร้าในระบบก่อน
	cart, err := s.cartRepo.GetCartByUserID(userID)
	if err != nil {
		return nil, errors.New("เกิดข้อผิดพลาดในการค้นหาตะกร้าสินค้า")
	}

	// 2. 💡 ท่าไม้ตาย "Lazy Initialization": ถ้าหาไม่เจอ แปลว่าเพิ่งสมัครสมาชิกใหม่ ให้สร้างตะกร้าให้เลย!
	if cart == nil {
		cart, err = s.cartRepo.CreateCart(userID)
		if err != nil {
			return nil, errors.New("ไม่สามารถสร้างตะกร้าสินค้าใหม่ได้")
		}
	}

	// 3. ไปกวาดของในตะกร้า (พร้อมรายละเอียดสินค้า) มาใส่ใน struct
	items, err := s.cartRepo.GetCartItemsWithDetails(cart.ID)
	if err != nil {
		return nil, errors.New("เกิดข้อผิดพลาดในการดึงรายการสินค้า")
	}

	cart.Items = items
	return cart, nil
}

// -----------------------------------------------------------------
// 2. หยิบของใส่ตะกร้า
// -----------------------------------------------------------------
func (s *cartService) AddToCart(userID uint, variantID uint, quantity int) error {
	// 🛡️ กฎข้อที่ 1: ห้ามหยิบของ 0 ชิ้น หรือติดลบ
	if quantity <= 0 {
		return errors.New("จำนวนสินค้าต้องมากกว่า 0")
	}

	// ไปเรียกฟังก์ชัน GetMyCart ด้านบน เพื่อเอา ID ตะกร้ามา (ถ้าไม่มีมันจะสร้างให้เองโดยอัตโนมัติ)
	cart, err := s.GetMyCart(userID)
	if err != nil {
		return err
	}

	/* 📝 หมายเหตุ: ในระบบ E-commerce ทั่วไป เรามักจะ "ปล่อยให้หยิบใส่ตะกร้าไปก่อน"
	   (Optimistic Cart) แล้วค่อยไปเช็คสต็อกแบบเป๊ะๆ 100% อีกทีตอนกดปุ่ม "ชำระเงิน (Checkout)"
	   เพื่อลดภาระ Database ครับ เราเลยจะเซฟลง DB ทันที
	*/

	// ส่ง ID ตะกร้า, รหัสสินค้า และจำนวน ไปให้ Repo จัดการต่อ (Repo เราใช้ท่า UPSERT ไว้แล้ว)
	return s.cartRepo.AddItem(cart.ID, variantID, quantity)
}

// -----------------------------------------------------------------
// 3. ปรับเปลี่ยนจำนวนของในตะกร้า (เช่น กดปุ่ม + / - หน้าเว็บ)
// -----------------------------------------------------------------
func (s *cartService) UpdateQuantity(userID uint, cartItemID uint, quantity int) error {
	// 🛡️ กฎ: ถ้าปรับจำนวนให้กลายเป็น 0 หรือติดลบ ให้เตะออก
	// (เพราะถ้าลูกค้าจะลบ ควรไปเรียก API RemoveFromCart แทน)
	if quantity <= 0 {
		return errors.New("จำนวนสินค้าต้องมากกว่า 0 (หากต้องการนำออก ให้กดปุ่มลบสินค้า)")
	}

	// สั่งอัปเดตจำนวน
	return s.cartRepo.UpdateItemQuantity(cartItemID, quantity)
}

// -----------------------------------------------------------------
// 4. เอาของออกจากตะกร้า
// -----------------------------------------------------------------
func (s *cartService) RemoveFromCart(userID uint, cartItemID uint) error {
	// สั่งลบ Item นั้นๆ ทิ้งไปเลย
	return s.cartRepo.RemoveItem(cartItemID)
}
