package handler

import (
	"simple-clothes-shop/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	service domain.ProductService
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

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

	// 🚀 แทรก c.UserContext()
	products, err := h.service.FetchWithFilter(c.UserContext(), categoryID, minPrice, maxPrice, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"page":  page,
		"limit": limit,
		"data":  products,
	})
}

func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	product, err := h.service.FetchByID(c.UserContext(), uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบสินค้า"})
	}
	return c.JSON(product)
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var product domain.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูล JSON ไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.service.CreateProduct(c.UserContext(), &product); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(product)
}

func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.service.RemoveProduct(c.UserContext(), uint(id)); err != nil {
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

	// 🚀 แทรก c.UserContext()
	if err := h.service.UpdateProduct(c.UserContext(), uint(id), &product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตสินค้าสำเร็จ"})
}

func (h *ProductHandler) DeleteVariant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	err = h.service.RemoveVariant(c.UserContext(), uint(id))
	if err != nil {
		if err.Error() == "ไม่พบ Variant นี้ในระบบ (ลบไม่สำเร็จ)" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ลบตัวเลือกสินค้า (Variant) สำเร็จ"})
}
