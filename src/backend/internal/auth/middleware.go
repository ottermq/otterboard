package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
)

func AuthMiddleware(sessions SessionStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			return common.HandlerError(c, common.ErrUnauthorized)
		}
		userID, err := sessions.Get(c.Context(), sessionID)
		if err != nil {
			return common.HandlerError(c, common.ErrUnauthorized)
		}
		c.Locals("userID", userID)
		return c.Next()
	}
}
