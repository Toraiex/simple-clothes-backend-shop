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

func (h *OrderHandler) Create(c *fiber.Ctx) error {

	userID, err := GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var input struct {
		Items []struct {
			ProductID uint `json:"product_id"`
			Quantity  int  `json:"quantity"`
		} `json:"items"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	if len(input.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "order ต้องมีอย่างน้อย 1 สินค้า",
		})
	}

	var items []domain.OrderItem
	for _, i := range input.Items {
		items = append(items, domain.OrderItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
		})
	}

	if err := h.service.CreateOrder(userID, items); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "order created"})
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	userID, err := GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	role, err := GetUserRole(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var order *domain.Order

	if role == "admin" {
		order, err = h.service.GetByID(uint(id))
	} else {
		order, err = h.service.GetByIDForUser(userID, uint(id))
	}

	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(order)
}

func (h *OrderHandler) GetMyOrders(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	orders, err := h.service.GetByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(orders)
}
func (h *OrderHandler) Cancel(c *fiber.Ctx) error {

	userID, err := GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	idParam := c.Params("id")
	orderID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	err = h.service.CancelOrder(userID, uint(orderID))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "order canceled"})
}
func (h *OrderHandler) AdminUpdateStatus(c *fiber.Ctx) error {

	orderIDParam := c.Params("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid order id"})
	}

	var input struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	err = h.service.AdminUpdateStatus(
		uint(orderID),
		domain.OrderStatus(input.Status),
	)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "order status updated"})
}
