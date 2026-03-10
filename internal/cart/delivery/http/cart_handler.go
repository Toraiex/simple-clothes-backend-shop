package http

import (
	"fmt"
	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/middleware"
	"simple-clothes-shop/pkg/utils" // 💡 1. Import utils ส่วนกลาง
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	cartUsecase domain.CartUsecase
}

func NewCartHandler(api fiber.Router, uc domain.CartUsecase, authMid fiber.Handler) {
	handler := &CartHandler{cartUsecase: uc}
	cartGroup := api.Group("/cart")

	cartGroup.Get("/", authMid, handler.GetMyCart)
	cartGroup.Post("/", authMid, handler.AddToCart)
	cartGroup.Patch("/:id", authMid, handler.UpdateQuantity)
	cartGroup.Delete("/:id", authMid, handler.RemoveFromCart)
}
func getUserID(c *fiber.Ctx) (uint, error) {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return 0, domain.ErrUnauthorized
	}
	return userID, nil
}
func (h *CartHandler) GetMyCart(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": "เซสชันไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"})
	}

	cart, err := h.cartUsecase.GetMyCart(c.UserContext(), userID)
	if err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	// 💡 ปรับ format ตอบกลับให้คลีนๆ
	return c.JSON(fiber.Map{
		"message": "ดึงข้อมูลตะกร้าสินค้าสำเร็จ",
		"data":    cart,
	})
}

type AddToCartRequest struct {
	VariantID uint `json:"variant_id"`
	Quantity  int  `json:"quantity"`
}

func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": "เซสชันไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"})
	}

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.cartUsecase.AddToCart(c.UserContext(), userID, req.VariantID, req.Quantity); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		/*if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}*/
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "เพิ่มสินค้าลงตะกร้าเรียบร้อยแล้ว"})
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
}

func (h *CartHandler) UpdateQuantity(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": "เซสชันไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"})
	}

	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		err = fmt.Errorf("ID รายการสินค้าไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	var req UpdateCartRequest
	if err := c.BodyParser(&req); err != nil {
		err = fmt.Errorf("รูปแบบข้อมูลไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.cartUsecase.UpdateQuantity(c.UserContext(), userID, uint(cartItemID), req.Quantity); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "อัปเดตจำนวนสินค้าเรียบร้อยแล้ว"})
}

func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": "เซสชันไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"})
	}

	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		err = fmt.Errorf("ID รายการสินค้าไม่ถูกต้อง: %w", domain.ErrBadParamInput)
		return c.Status(utils.GetStatusCode(err)).JSON(fiber.Map{"message": err.Error()})
	}

	if err := h.cartUsecase.RemoveFromCart(c.UserContext(), userID, uint(cartItemID)); err != nil {
		statusCode := utils.GetStatusCode(err)
		errMsg := err.Error()
		if statusCode == fiber.StatusInternalServerError {
			errMsg = domain.ErrInternalServerError.Error()
		}
		return c.Status(statusCode).JSON(fiber.Map{"message": errMsg})
	}

	return c.JSON(fiber.Map{"message": "ลบสินค้าออกจากตะกร้าเรียบร้อยแล้ว"})
}
