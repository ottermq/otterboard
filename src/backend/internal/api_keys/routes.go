package api_keys

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterApiKeyRoutes(wsGroup fiber.Router, h *Handler) {
	g := wsGroup.Group("/api-keys", middleware.RequireRole(middleware.RoleAdmin))
	g.Post("/", h.CreateApiKey)
	g.Get("/", h.ListApiKeys)
	g.Delete("/:keyId", h.RevokeApiKey)
}
