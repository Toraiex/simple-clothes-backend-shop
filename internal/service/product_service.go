package service

import (
	"errors"
	"simple-clothes-shop/internal/domain" // เรียกใช้กฎจาก Domain
)

// สร้าง Struct สำหรับ Service
type productService struct {
	repo         domain.ProductRepository
	categoryRepo domain.CategoryRepository // 👈 เพิ่มฟิลด์นี้
}

// 👈 แก้ไขให้รับ 2 parameters
func NewProductService(repo domain.ProductRepository, catRepo domain.CategoryRepository) domain.ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: catRepo,
	}
}

func (s *productService) UpdateProduct(id uint, product *domain.Product) error {
	// 1. เช็คก่อนว่าหมวดหมู่ที่ส่งมามีจริงไหม
	if product.CategoryID != 0 {
		_, err := s.categoryRepo.GetByID(product.CategoryID)
		if err != nil {
			return errors.New("ไม่พบหมวดหมู่สินค้าที่ระบุ")
		}
	}

	// 2. สั่งอัปเดต
	return s.repo.Update(id, product)
}

// ==========================================
// เริ่มเขียน Logic (Implement Service Interface)
// ==========================================

// 1. ดึงสินค้าทั้งหมด
func (s *productService) FetchAll() ([]domain.Product, error) {
	// ตรงนี้เราสั่งให้ Repo ไปหยิบของมาได้เลย
	return s.repo.GetAll()
}

// 2. ดึงสินค้าตาม ID
func (s *productService) FetchByID(id uint) (*domain.Product, error) {
	return s.repo.GetByID(id)
}

// 3. สร้างสินค้าใหม่ (ที่มีการเช็คกฎธุรกิจ)
func (s *productService) CreateProduct(product *domain.Product) error {
	// 🛡️ ตัวอย่าง Business Logic 1: ห้ามตั้งราคาติดลบ
	if product.Price <= 0 {
		return errors.New("ราคาพื้นฐานของสินค้าต้องมากกว่า 0 บาท")
	}

	// 🛡️ ตัวอย่าง Business Logic 2: ถ้าเป็นเสื้อผ้า ต้องมีสต็อกอย่างน้อย 1 ชิ้น
	if product.Stock < 0 {
		return errors.New("สต็อกสินค้าไม่สามารถติดลบได้")
	}

	// 🛡️ ตัวอย่าง Business Logic 3: ตรวจสอบความถูกต้องของ Variants
	for _, v := range product.Variants {
		if v.Price <= 0 {
			return errors.New("ราคาของสินค้าแต่ละสี/ไซส์ (Variant) ต้องมากกว่า 0 บาท")
		}
	}

	// ถ้าผ่านกฎทุกอย่างแล้ว ถึงจะอนุญาตให้ Repo บันทึกลง DB
	return s.repo.Create(product)
}

// 4. ลบสินค้า
func (s *productService) RemoveProduct(id uint) error {
	// อาจจะเพิ่ม Logic เช็คว่าสินค้านี้มียอดค้างส่งไหมก่อนลบก็ได้
	return s.repo.Delete(id)
}
