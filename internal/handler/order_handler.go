package handler

import (
	"simple-clothes-shop/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	service domain.OrderService
}

func NewOrderHandler(service domain.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.service.Checkout(c.UserContext(), userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "สร้างคำสั่งซื้อสำเร็จและล้างตะกร้าเรียบร้อยแล้ว",
	})
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	idParam, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID คำสั่งซื้อไม่ถูกต้อง"})
	}

	userID, ok := c.Locals("user_id").(uint)
	role, roleOk := c.Locals("role").(string)
	if !ok || !roleOk {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var order *domain.Order
	// 🚀 แทรก c.UserContext()
	if role == "admin" {
		order, err = h.service.GetByID(c.UserContext(), uint(idParam))
	} else {
		order, err = h.service.GetByIDForUser(c.UserContext(), userID, uint(idParam))
	}

	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(order)
}

func (h *OrderHandler) GetMyOrders(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	// 🚀 แทรก c.UserContext()
	orders, err := h.service.GetByUserID(c.UserContext(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID คำสั่งซื้อไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.service.CancelOrder(c.UserContext(), userID, uint(orderID)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ยกเลิกคำสั่งซื้อและคืนสต็อกเรียบร้อยแล้ว"})
}

func (h *OrderHandler) AdminUpdateStatus(c *fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID คำสั่งซื้อไม่ถูกต้อง"})
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
	}

	// 🚀 แทรก c.UserContext()
	if err := h.service.AdminUpdateStatus(c.UserContext(), uint(orderID), domain.OrderStatus(input.Status)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "อัปเดตสถานะคำสั่งซื้อเรียบร้อยแล้ว"})
}
