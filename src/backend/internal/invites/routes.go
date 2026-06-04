package invites

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterInviteRoutes(api fiber.Router, h *Handler) {
	api.Get("/invites/:token", h.GetInvite)
}

func RegisterProtectedInviteRoutes(api fiber.Router, h *Handler) {
	api.Post("/invites/:token/accept", h.AcceptInvite)
}

func RegisterWorkspaceScopedInviteRoutes(wsGroup fiber.Router, h *Handler) {
	wsGroup.Post("/invites", middleware.RequireRole(middleware.RoleAdmin), h.GenerateInvite)
}
