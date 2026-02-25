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
	// ดึงค่าจาก key "user_id" ที่ AuthMiddleware ส่งมาให้
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	return userID, nil
}

// -----------------------------------------------------------------
// 1. [GET] ดึงข้อมูลตะกร้าของฉัน
// -----------------------------------------------------------------
func (h *CartHandler) GetMyCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	cart, err := h.cartService.GetMyCart(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ดึงข้อมูลตะกร้าสินค้าสำเร็จ",
		"data":    cart,
	})
}

// -----------------------------------------------------------------
// 2. [POST] หยิบของใส่ตะกร้า (รับ JSON: variant_id, quantity)
// -----------------------------------------------------------------
type AddToCartRequest struct {
	VariantID uint `json:"variant_id"`
	Quantity  int  `json:"quantity"`
}

func (h *CartHandler) AddToCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	// 1. รับค่า JSON จาก Body
	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// 2. ส่งให้ Service จัดการ
	if err := h.cartService.AddToCart(userID, req.VariantID, req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "เพิ่มสินค้าลงตะกร้าเรียบร้อยแล้ว",
	})
}

// -----------------------------------------------------------------
// 3. [PATCH] ปรับจำนวนของในตะกร้า
// -----------------------------------------------------------------
type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
}

func (h *CartHandler) UpdateQuantity(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	// ดึง cart_item_id จาก URL (เช่น /api/cart/items/5)
	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID รายการสินค้าไม่ถูกต้อง"})
	}

	var req UpdateCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	if err := h.cartService.UpdateQuantity(userID, uint(cartItemID), req.Quantity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "อัปเดตจำนวนสินค้าเรียบร้อยแล้ว",
	})
}

// -----------------------------------------------------------------
// 4. [DELETE] ลบของออกจากตะกร้า
// -----------------------------------------------------------------
func (h *CartHandler) RemoveFromCart(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	cartItemID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID รายการสินค้าไม่ถูกต้อง"})
	}

	if err := h.cartService.RemoveFromCart(userID, uint(cartItemID)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ลบสินค้าออกจากตะกร้าเรียบร้อยแล้ว",
	})
}
