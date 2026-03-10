package http

import (
	"fmt"
	"strconv"

	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	usecase domain.CategoryUsecase
}

func NewCategoryHandler(api fiber.Router, uc domain.CategoryUsecase, authMid fiber.Handler, adminMid fiber.Handler) {
	handler := &CategoryHandler{usecase: uc}

	catGroup := api.Group("/categories")
	catGroup.Get("/", handler.GetAll)
	catGroup.Get("/:id", handler.GetByID)

	catGroup.Post("/", authMid, adminMid, handler.Create)
	catGroup.Put("/:id", authMid, adminMid, handler.Update)
	catGroup.Delete("/:id", authMid, adminMid, handler.Delete)
}

func (h *CategoryHandler) GetAll(c *fiber.Ctx) error {
	categories, err := h.usecase.FetchAll(c.UserContext())
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}
	return c.JSON(categories)
}

func (h *CategoryHandler) Create(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.CreateCategory(c.UserContext(), input.Name); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "สร้างหมวดหมู่สำเร็จ"})
}

func (h *CategoryHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID หมวดหมู่ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	category, err := h.usecase.GetCategory(c.UserContext(), uint(id))
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}
	return c.JSON(category)
}

func (h *CategoryHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID หมวดหมู่ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.UpdateCategory(c.UserContext(), uint(id), input.Name); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "แก้ไขหมวดหมู่สำเร็จ"})
}

func (h *CategoryHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		err = fmt.Errorf("ID หมวดหมู่ไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.usecase.RemoveCategory(c.UserContext(), uint(id)); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "ลบหมวดหมู่สำเร็จ"})
}
