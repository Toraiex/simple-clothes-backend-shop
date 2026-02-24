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

// =================================================================
// 🛒 1. [POST] ยืนยันคำสั่งซื้อ (Checkout) จากตะกร้า
// =================================================================
func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	// ดึง ID จาก Middleware (แบบเดียวกับที่เราทำใน Cart)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	// สั่ง Checkout ได้เลย! (ไม่ต้องรับ Body JSON อะไรทั้งนั้น หน้าบ้านยิง POST เปล่าๆ มาได้เลย)
	if err := h.service.Checkout(userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "สร้างคำสั่งซื้อสำเร็จและล้างตะกร้าเรียบร้อยแล้ว",
	})
}

// =================================================================
// 2. [GET] ดูรายละเอียดบิล
// =================================================================
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
	if role == "admin" {
		order, err = h.service.GetByID(uint(idParam))
	} else {
		order, err = h.service.GetByIDForUser(userID, uint(idParam))
	}

	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(order)
}

// =================================================================
// 3. [GET] ดูประวัติการสั่งซื้อของฉัน
// =================================================================
func (h *OrderHandler) GetMyOrders(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	orders, err := h.service.GetByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

// =================================================================
// 4. [PUT] ลูกค้ายกเลิกคำสั่งซื้อ
// =================================================================
func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID คำสั่งซื้อไม่ถูกต้อง"})
	}

	if err := h.service.CancelOrder(userID, uint(orderID)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ยกเลิกคำสั่งซื้อและคืนสต็อกเรียบร้อยแล้ว"})
}

// =================================================================
// 5. [PUT] แอดมินอัปเดตสถานะ (Paid, Shipped, ฯลฯ)
// =================================================================
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

	if err := h.service.AdminUpdateStatus(uint(orderID), domain.OrderStatus(input.Status)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "อัปเดตสถานะคำสั่งซื้อเรียบร้อยแล้ว"})
}
