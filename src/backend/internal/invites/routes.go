package invites

import "github.com/gofiber/fiber/v2"

func RegisterInviteRoutes(api fiber.Router, h *Handler) {
	api.Get("/invites/:token", h.GetInvite)
}

func RegisterProtectedInviteRoutes(api fiber.Router, h *Handler) {
	api.Post("/workspaces/:workspaceId/invites", h.GenerateInvite)
	api.Post("/invites/:token/accept", h.AcceptInvite)
}
