package api_keys

import "github.com/gofiber/fiber/v2"

func RegisterApiKeyRoutes(api fiber.Router, h *Handler) {
	g := api.Group("/workspaces")
	g.Post("/:workspaceId/api-keys", h.CreateApiKey)
	g.Get("/:workspaceId/api-keys", h.ListApiKeys)
	g.Delete("/:workspaceId/api-keys/:keyId", h.RevokeApiKey)
}
