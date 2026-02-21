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
	// 1. FETCH: ดึงข้อมูลสินค้า "ตัวเก่า" จาก Database ออกมาก่อน
	existingProduct, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("ไม่พบสินค้าที่ต้องการแก้ไข")
	}

	// 2. MERGE: เอาข้อมูลใหม่ไป "แปะทับ" ข้อมูลเก่า (เฉพาะอันที่ส่งมา)

	// ถ้ามีการส่งชื่อมาใหม่ (ไม่ใช่ค่าว่าง) ให้ใช้ชื่อใหม่
	if product.Name != "" {
		existingProduct.Name = product.Name
	}
	// ถ้ามีการส่งรายละเอียดมาใหม่
	if product.Description != "" {
		existingProduct.Description = product.Description
	}
	// ถ้าส่งราคามา (และมากกว่า 0)
	if product.Price > 0 {
		existingProduct.Price = product.Price
	}
	// ถ้าส่งสต็อกมา (สต็อกเป็น 0 ได้ ต้องระวัง logic นี้ถ้าอยากให้แก้เป็น 0 ได้จริงๆ อาจต้องใช้ Pointer แต่เบื้องต้นใช้แบบนี้ก่อนได้)
	if product.Stock >= 0 {
		existingProduct.Stock = product.Stock
	}
	// ถ้าเปลี่ยนหมวดหมู่
	if product.CategoryID != 0 {
		// เช็คก่อนว่าหมวดหมู่ใหม่มีจริงไหม
		_, err := s.categoryRepo.GetByID(product.CategoryID)
		if err != nil {
			return errors.New("ไม่พบหมวดหมู่สินค้าที่ระบุ")
		}
		existingProduct.CategoryID = product.CategoryID
	}
	// ถ้าเปลี่ยนรูป
	if len(product.Images) > 0 {
		existingProduct.Images = product.Images
	}

	// 3. VARIANTS: ส่ง Variants ใหม่ไปให้ Repo จัดการต่อ (Repo เราเขียน Logic Upsert ไว้แล้ว)
	// แต่เราต้องแนบ Variants ที่ส่งมาใหม่ เข้าไปใน existingProduct
	if len(product.Variants) > 0 {
		existingProduct.Variants = product.Variants
	}

	// 4. SAVE: ส่งข้อมูลที่ "ผสานร่างเสร็จแล้ว" กลับไปบันทึก
	// (Repo จะมองว่าเป็น PUT เหมือนเดิม คือเซฟทับทุกช่อง แต่ตอนนี้ข้อมูลเราครบแล้ว)
	return s.repo.Update(id, existingProduct)
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
func (s *productService) FetchByCategoryID(categoryID uint) ([]domain.Product, error) {
	return s.repo.GetByCategoryID(categoryID)
}
func (s *productService) FetchWithFilter(
	categoryID *uint,
	minPrice *float64,
	maxPrice *float64,
) ([]domain.Product, error) {

	return s.repo.GetWithFilter(categoryID, minPrice, maxPrice)
}

// ... (ฟังก์ชันอื่นเดิม) ...

// ✅ เพิ่มฟังก์ชันนี้ลงไปท้ายไฟล์
func (s *productService) RemoveVariant(variantID uint) error {
	return s.repo.DeleteVariant(variantID)
}
