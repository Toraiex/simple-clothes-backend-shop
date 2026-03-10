package http

import (
	"simple-clothes-shop/internal/domain"
	"simple-clothes-shop/internal/middleware" // 💡 Import middleware เข้ามาใช้งาน
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	usecase domain.OrderUsecase
}

func NewOrderHandler(router fiber.Router, service domain.OrderUsecase, authMid fiber.Handler, adminMid fiber.Handler) {
	handler := &OrderHandler{usecase: service}

	orderGroup := router.Group("/orders")
	orderGroup.Post("/checkout", authMid, handler.Checkout)
	orderGroup.Get("/me", authMid, handler.GetMyOrders)
	orderGroup.Get("/:id", authMid, handler.GetByID)
	orderGroup.Patch("/:id/cancel", authMid, handler.Cancel)
	orderGroup.Patch("/:id/status", authMid, adminMid, handler.AdminUpdateStatus)
}

func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	// 💡 ใช้ Helper แทน c.Locals
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	if err := h.usecase.Checkout(c.UserContext(), userID); err != nil {
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

	// 💡 ดึง ID และ Role อย่างปลอดภัย
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	role, err := middleware.GetUserRole(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	var order *domain.Order
	if role == string(domain.RoleAdmin) { // หรือ == "admin" ก็ได้
		order, err = h.usecase.GetByID(c.UserContext(), uint(idParam))
	} else {
		order, err = h.usecase.GetByIDForUser(c.UserContext(), userID, uint(idParam))
	}

	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(order)
}

func (h *OrderHandler) GetMyOrders(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	orders, err := h.usecase.GetByUserID(c.UserContext(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "กรุณาเข้าสู่ระบบ"})
	}

	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID คำสั่งซื้อไม่ถูกต้อง"})
	}

	if err := h.usecase.CancelOrder(c.UserContext(), userID, uint(orderID)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "ยกเลิกคำสั่งซื้อและคืนสต็อกเรียบร้อยแล้ว"})
}

func (h *OrderHandler) AdminUpdateStatus(c *fiber.Ctx) error {
	// ไม่ต้องดึง userID เพราะมี AdminMiddleware กันไว้อยู่แล้ว
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

	if err := h.usecase.AdminUpdateStatus(c.UserContext(), uint(orderID), domain.OrderStatus(input.Status)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "อัปเดตสถานะคำสั่งซื้อเรียบร้อยแล้ว"})
}
