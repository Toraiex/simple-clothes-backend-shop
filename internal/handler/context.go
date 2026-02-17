package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

const (
	ContextUserIDKey = "userID"
	ContextRoleKey   = "role"
)

func SetUserContext(c *fiber.Ctx, userID uint, role string) {
	c.Locals(ContextUserIDKey, userID)
	c.Locals(ContextRoleKey, role)
}

func GetUserID(c *fiber.Ctx) (uint, error) {
	v := c.Locals(ContextUserIDKey)

	userID, ok := v.(uint)
	if !ok {
		return 0, errors.New("invalid user id in context")
	}

	return userID, nil
}

func GetUserRole(c *fiber.Ctx) (string, error) {
	v := c.Locals(ContextRoleKey)

	role, ok := v.(string)
	if !ok {
		return "", errors.New("invalid role in context")
	}

	return role, nil
}
