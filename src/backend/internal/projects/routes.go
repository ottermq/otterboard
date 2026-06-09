package projects

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/middleware"
)

func RegisterProjectRoutes(wsGroup fiber.Router, h *Handler) {
	g := wsGroup.Group("/projects")
	g.Get("/", h.ListProjects)
	g.Get("/:id", h.GetProject)
	g.Post("/", middleware.RequireRole(
		middleware.RoleAdmin,
		middleware.RoleMember),
		h.CreateProject)
	g.Patch("/:id", middleware.RequireRole(
		middleware.RoleAdmin,
		middleware.RoleMember),
		h.UpdateProject)
	g.Delete("/:id", middleware.RequireRole(
		middleware.RoleAdmin),
		h.DeleteProject)
}
