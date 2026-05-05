package common

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func HandlerError(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{"error": appErr.Message})
	}
	log.Printf("unexpected error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
}
