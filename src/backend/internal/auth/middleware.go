package auth

import (
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(sessions SessionStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		userID, err := sessions.Get(c.Context(), sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		c.Locals("userID", userID)
		return c.Next()
	}
}
