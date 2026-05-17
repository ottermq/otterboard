package routes

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) fiber.Router {
	g := app.Group("/api/v1")
	g.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
	return g
}

func RegisterProtectedRoutes(app *fiber.App, authMiddleware fiber.Handler) fiber.Router {
	return app.Group("/api/v1", authMiddleware)
}
