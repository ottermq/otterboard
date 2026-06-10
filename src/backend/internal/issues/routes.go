package issues

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterIssueRoutes(wsGroup fiber.Router, h *Handler) {
	g := wsGroup.Group("/projects/:projectId/issues")
	g.Get("/", h.ListIssues)
	g.Post("/", middleware.RequireRole(
		middleware.RoleMember,
		middleware.RoleAdmin),
		h.CreateIssue)
	g.Get("/:id", h.GetIssue)
	g.Patch("/:id", middleware.RequireRole(
		middleware.RoleMember,
		middleware.RoleAdmin),
		h.UpdateIssue)
	g.Delete("/:id", middleware.RequireRole(
		middleware.RoleAdmin),
		h.DeleteIssue)
	wsGroup.Get("/issues", h.ListIssuesByWorkspace)
}
