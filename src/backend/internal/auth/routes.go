package auth

import "github.com/gofiber/fiber/v2"

func RegisterAuthRoutes(app *fiber.App, h *Handler) {
	g := app.Group("/api/v1/auth")
	g.Post("/register", h.Register)
}
