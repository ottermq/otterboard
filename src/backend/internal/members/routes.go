package members

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterMemberRoutes(wsGroup fiber.Router, h *Handler) {
	g := wsGroup.Group("/members")
	g.Get("/", h.ListMembers)

	adminGroup := g.Group("", middleware.RequireRole(middleware.RoleAdmin))
	adminGroup.Patch("/:userId", h.UpdateMemberRole)
	adminGroup.Delete("/:userId", h.RemoveMember)
}
