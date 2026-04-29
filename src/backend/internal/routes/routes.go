package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/auth"
)

var (
	v1 = "/api/v1"
)

func RegisterRoutes(app *fiber.App) {
	// API v1 routes
	apiV1 := app.Group(v1)
	apiV1.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
}

func RegisterAuthRoutes(app *fiber.App, authHandler *auth.Handler) {
	apiV1 := app.Group(v1 + "/auth")
	apiV1.Post("/register", authHandler.Register)
}
