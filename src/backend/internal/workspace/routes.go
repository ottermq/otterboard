package workspace

import "github.com/gofiber/fiber/v2"

func RegisterWorkspacesRoutes(api fiber.Router, h *Handler) {
	g := api.Group("/workspaces")
	g.Post("/", h.CreateWorkspace)
	g.Get("/", h.ListWorkspaces)
	g.Get("/:id", h.GetWorkspace)
	g.Patch("/:id", h.UpdateWorkspace)
	g.Delete("/:id", h.DeleteWorkspace)
}
