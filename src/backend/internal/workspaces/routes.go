package workspaces

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterWorkspacesRoutes(api fiber.Router, h *Handler) {
	g := api.Group("/workspaces")
	g.Post("/", h.CreateWorkspace)
	g.Get("/", h.ListWorkspaces)
}

func RegisterWorkspacesScopedRoutes(wsGroup fiber.Router, h *Handler) {
	wsGroup.Get("", h.GetWorkspace)

	adminGroup := wsGroup.Group("", middleware.RequireRole(middleware.RoleAdmin))
	adminGroup.Patch("", h.UpdateWorkspace)
	adminGroup.Delete("", h.DeleteWorkspace)
}
