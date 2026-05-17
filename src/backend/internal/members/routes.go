package members

import "github.com/gofiber/fiber/v2"

func RegisterMemberRoutes(api fiber.Router, h *Handler) {
	g := api.Group("/workspaces/:workspaceId/members")
	g.Get("/", h.ListMembers)
	g.Patch("/:userId", h.UpdateMemberRole)
	g.Delete("/:userId", h.RemoveMember)
}
