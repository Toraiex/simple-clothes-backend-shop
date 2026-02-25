package handler

import (
	"simple-clothes-shop/internal/domain" // เรียกใช้กฎจาก Domain
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// สร้าง Struct สำหรับ Handler
type ProductHandler struct {
	service domain.ProductService // ถือ Interface ของ Service ไว้
}

// ฟังก์ชันสร้าง Handler ใหม่ (Constructor)
func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// ==========================================
// เริ่มทำงาน (Handler Methods)
// ==========================================

// 🟢 ดึงสินค้าทั้งหมด
func (h *ProductHandler) GetAll(c *fiber.Ctx) error {

	var categoryID *uint
	var minPrice *float64
	var maxPrice *float64

	if v := c.Query("category_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid category_id"})
		}
		temp := uint(id)
		categoryID = &temp
	}

	if v := c.Query("min_price"); v != "" {
		price, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid min_price"})
		}
		minPrice = &price
	}

	if v := c.Query("max_price"); v != "" {
		price, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid max_price"})
		}
		maxPrice = &price
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// เรียก Service
	products, err := h.service.FetchWithFilter(categoryID, minPrice, maxPrice, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"page":  page,
		"limit": limit,
		"data":  products,
	})
}

// 🟢 ดึงสินค้าตาม ID
func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	product, err := h.service.FetchByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบสินค้า"})
	}
	return c.JSON(product)
}

// 🔒 สร้างสินค้าใหม่
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var product domain.Product
	// 1. แปลง JSON Body ให้กลายเป็น Struct Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูล JSON ไม่ถูกต้อง"})
	}

	// 2. ส่งให้ Service (สมอง) เป็นคนจัดการต่อ
	if err := h.service.CreateProduct(&product); err != nil {
		// ถ้า Service ตรวจแล้วติด Error (เช่น ราคาติดลบ) จะส่งกลับมาที่นี่
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. ถ้าสำเร็จ ส่ง 201 Created กลับไป
	return c.Status(201).JSON(product)
}

// 🔒 ลบสินค้า
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	if err := h.service.RemoveProduct(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถลบสินค้าได้"})
	}
	return c.JSON(fiber.Map{"message": "ลบสินค้าเรียบร้อยแล้ว"})
}
func (h *ProductHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var product domain.Product

	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if err := h.service.UpdateProduct(uint(id), &product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตสินค้าสำเร็จ"})
}

// 🔒 ลบ Variant (สี/ไซส์)
func (h *ProductHandler) DeleteVariant(c *fiber.Ctx) error {
	// 1. ดึง ID จาก URL parameter
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	// 2. เรียก Service ให้ลบ Variant
	err = h.service.RemoveVariant(uint(id))
	if err != nil {
		// เช็ค Error ถ้าหาไม่เจอ
		if err.Error() == "ไม่พบ Variant นี้ในระบบ (ลบไม่สำเร็จ)" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ลบตัวเลือกสินค้า (Variant) สำเร็จ"})
}
