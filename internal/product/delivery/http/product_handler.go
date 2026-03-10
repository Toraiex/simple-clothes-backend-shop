package http

import (
	// 👈 สำคัญมาก อย่าลืม import ตัวนี้ครับ
	"fmt"
	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// 💡 1. เอา Helper ตัวเก่งของเรามาแปะไว้ท้ายไฟล์ หรือจะ import จาก pkg ที่เราเคยคุยกันก็ได้ครับ
type ProductHandler struct {
	usecase domain.ProductUsecase
}

func NewProductHandler(router fiber.Router, usecase domain.ProductUsecase, authMid fiber.Handler, adminMid fiber.Handler) {
	handler := &ProductHandler{usecase: usecase}

	productGroup := router.Group("/products")

	productGroup.Get("/", handler.GetAll)
	productGroup.Get("/:id", handler.GetByID)

	productGroup.Post("/", authMid, adminMid, handler.Create)
	productGroup.Patch("/:id", authMid, adminMid, handler.Update)
	productGroup.Delete("/:id", authMid, adminMid, handler.Delete)
	productGroup.Delete("/variants/:id", authMid, adminMid, handler.DeleteVariant)
}

func (h *ProductHandler) GetAll(c *fiber.Ctx) error {
	var categoryID *uint
	var minPrice *float64
	var maxPrice *float64

	if v := c.Query("category_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			err = fmt.Errorf("รูปแบบหมวดหมู่ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
			return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()}) // 💡 2. ใช้ utils.GetStatusCode
		}
		temp := uint(id)
		categoryID = &temp
	}

	if v := c.Query("min_price"); v != "" {
		price, err := strconv.ParseFloat(v, 64)
		if err != nil {
			err = fmt.Errorf("รูปแบบราคาไม่ถูกต้อง: %w", domain.ErrBadParamInput)
			return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
		}
		minPrice = &price
	}

	if v := c.Query("max_price"); v != "" {
		price, err := strconv.ParseFloat(v, 64)
		if err != nil {
			err = fmt.Errorf("รูปแบบราคาไม่ถูกต้อง: %w", domain.ErrBadParamInput)
			return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
		}
		maxPrice = &price
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	products, err := h.usecase.FetchWithFilter(c.UserContext(), categoryID, minPrice, maxPrice, page, limit)
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		} // 💡 ดัก DB หลุด
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"page": page, "limit": limit, "data": products})
}

func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID สินค้าไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	product, err := h.usecase.FetchByID(c.UserContext(), uint(id))
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}
	return c.JSON(product)
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var product domain.Product
	if err := c.BodyParser(&product); err != nil {
		err = fmt.Errorf("ข้อมูล JSON ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.CreateProduct(c.UserContext(), &product); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.Status(201).JSON(product)
}

func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID สินค้าไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.RemoveProduct(c.UserContext(), uint(id)); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}
	return c.JSON(fiber.Map{"message": "ลบสินค้าเรียบร้อยแล้ว"})
}

func (h *ProductHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var product domain.Product

	if err := c.BodyParser(&product); err != nil {
		err = fmt.Errorf("ข้อมูล JSON ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.UpdateProduct(c.UserContext(), uint(id), &product); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตสินค้าสำเร็จ"})
}

func (h *ProductHandler) DeleteVariant(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID ตัวเลือกสินค้าไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	err = h.usecase.RemoveVariant(c.UserContext(), uint(id))
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "ลบตัวเลือกสินค้า (Variant) สำเร็จ"})
}
