package http

import (
	"strconv"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	service domain.CategoryUsecase
}

func NewCategoryHandler(service domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	categories, err := h.usecase.FetchAll(c.UserContext()) // 👈 ใส่ c.UserContext()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(categories)
}

// internal/handler/category_handler.go
func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&input); err != nil {
		// 🚀 ใช้ Standard Response
		return response.Error(c, fiber.StatusBadRequest, "รูปแบบข้อมูลไม่ถูกต้อง", err.Error())
	}

	if err := h.usecase.CreateCategory(c.UserContext(), input.Name); err != nil {
		// 🚀 ใช้ Standard Response
		return response.Error(c, fiber.StatusUnprocessableEntity, "สร้างหมวดหมู่ไม่สำเร็จ", err.Error())
	}

	// 🚀 ใช้ Standard Response
	return response.Success(c, fiber.StatusCreated, "สร้างหมวดหมู่สำเร็จ", nil)
}

func (h *CategoryHandler) GetByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	category, err := h.usecase.GetCategory(c.UserContext(), uint(id)) // 👈 ใส่ c.UserContext()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบหมวดหมู่นี้"})
	}
	return c.JSON(category)
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID หมวดหมู่ไม่ถูกต้อง"})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	if err := h.usecase.UpdateCategory(c.UserContext(), uint(id), input.Name); err != nil { // 👈 ใส่ c.UserContext()
		if err.Error() == "ไม่พบหมวดหมู่นี้ในระบบ" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "แก้ไขหมวดหมู่สำเร็จ"})
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID หมวดหมู่ไม่ถูกต้อง"})
	}

	if err := h.usecase.RemoveCategory(c.UserContext(), uint(id)); err != nil { // 👈 ใส่ c.UserContext()
		if err.Error() == "ไม่พบข้อมูลหมวดหมู่นี้ในระบบ (ลบไม่สำเร็จ)" {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "ลบหมวดหมู่สำเร็จ"})
}
