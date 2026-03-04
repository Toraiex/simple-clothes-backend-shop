package http

import (
	"simple-clothes-shop/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	cartUsecase domain.CartUsecase
}

// 🚀 เพิ่ม Parameters ให้ครบตามที่ router.go ส่งมา และให้มันผูก Route เองเลย
func NewCartHandler(api fiber.Router, uc domain.CartUsecase, authMid fiber.Handler) {
	handler := &CartHandler{cartUsecase: uc}

	// 🚀 สร้าง Group สำหรับ Cart
	cartGroup := api.Group("/cart")

	// 🚀 เอา Route มากางไว้ที่นี่ และใส่ Middleware เข้าไป
	cartGroup.Get("/", authMid, handler.GetMyCart)
	cartGroup.Post("/", authMid, handler.AddToCart)
	cartGroup.Patch("/:id", authMid, handler.UpdateQuantity)
	cartGroup.Delete("/:id", authMid, handler.RemoveFromCart)
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
	cart, err := h.cartUsecase.GetMyCart(c.UserContext(), userID)
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
	if err := h.cartUsecase.AddToCart(c.UserContext(), userID, req.VariantID, req.Quantity); err != nil {
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
	if err := h.cartUsecase.UpdateQuantity(c.UserContext(), userID, uint(cartItemID), req.Quantity); err != nil {
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
	if err := h.cartUsecase.RemoveFromCart(c.UserContext(), userID, uint(cartItemID)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ลบสินค้าออกจากตะกร้าเรียบร้อยแล้ว",
	})
}
