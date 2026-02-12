package handler

import (
	"strconv"

	"simple-clothes-shop/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	service domain.CategoryService
}

func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	categories, err := h.service.FetchAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if err := h.service.CreateCategory(input.Name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "สร้างหมวดหมู่สำเร็จ"})
}

func (h *CategoryHandler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	category, err := h.service.GetCategory(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบหมวดหมู่นี้"})
	}
	return c.JSON(category)
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if err := h.service.UpdateCategory(uint(id), input.Name); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "แก้ไขหมวดหมู่สำเร็จ"})
}
func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ไม่ถูกต้อง"})
	}

	if err := h.service.RemoveCategory(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ลบหมวดหมู่สำเร็จ"})
}
