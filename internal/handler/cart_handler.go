package handler

import (
	"simple-clothes-shop/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	cartService domain.CartService
}

func NewCartHandler(cartService domain.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func getUserID(c *fiber.Ctx) (uint, error) {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	return userID, nil
}

func (h *CartHandler) GetMyCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	// 🚀 แทรก c.UserContext()
	cart, err := h.cartService.GetMyCart(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ดึงข้อมูลตะกร้าสินค้าสำเร็จ",
		"data":    cart,
	})
}

type AddToCartRequest struct {
	VariantID uint `json:"variant_id"`
	Quantity  int  `json:"quantity"`
}

func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.cartService.AddToCart(c.UserContext(), userID, req.VariantID, req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "เพิ่มสินค้าลงตะกร้าเรียบร้อยแล้ว",
	})
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
}

func (h *CartHandler) UpdateQuantity(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID รายการสินค้าไม่ถูกต้อง"})
	}

	var req UpdateCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.cartService.UpdateQuantity(c.UserContext(), userID, uint(cartItemID), req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "อัปเดตจำนวนสินค้าเรียบร้อยแล้ว",
	})
}

func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID รายการสินค้าไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.cartService.RemoveFromCart(c.UserContext(), userID, uint(cartItemID)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ลบสินค้าออกจากตะกร้าเรียบร้อยแล้ว",
	})
}
