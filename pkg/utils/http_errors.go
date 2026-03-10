package utils

import (
	"errors"
	"simple-clothes-shop/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// GetStatusCode เป็นฟังก์ชันส่วนกลางสำหรับแปลง Domain Error เป็น HTTP Status Code
func GetStatusCode(err error) int {
	if err == nil {
		return fiber.StatusOK
	}

	switch {
	case errors.Is(err, domain.ErrInternalServerError):
		return fiber.StatusInternalServerError
	case errors.Is(err, domain.ErrNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrBadParamInput):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrUnauthorized):
		return fiber.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden
	case errors.Is(err, domain.ErrTooManyRequests):
		return fiber.StatusTooManyRequests
	default:
		return fiber.StatusInternalServerError
	}
}
