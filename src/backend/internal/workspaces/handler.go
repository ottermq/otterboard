package workspaces

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *WorkspaceService
}

func NewHandler(service *WorkspaceService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateWorkspace(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.HandlerError(c, common.ErrUnauthorized)
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	workspace, err := h.service.CreateWorkspace(c.Context(), CreateWorkspaceInput{
		Name:    req.Name,
		OwnerID: userID,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(mapToWorkspaceDto(workspace))
}

func (h *Handler) GetWorkspace(c *fiber.Ctx) error {
	workspace, err := h.service.GetWorkspaceByID(c.Context(), GetWorkspaceByIdInput{
		ID: c.Params("id"),
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToWorkspaceDto(workspace))
}

func (h *Handler) ListWorkspaces(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.HandlerError(c, common.ErrUnauthorized)
	}

	workspaces, err := h.service.GetWorkspacesByMemberID(c.Context(), userID)
	if err != nil {
		return common.HandlerError(c, err)
	}

	response := make([]dtos.WorkspaceDto, len(workspaces))
	for i, workspace := range workspaces {
		response[i] = mapToWorkspaceDto(workspace)
	}
	return c.JSON(response)
}

func (h *Handler) UpdateWorkspace(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	workspace, err := h.service.UpdateWorkspace(c.Context(), UpdateWorkspaceInput{
		ID:   c.Params("id"),
		Name: req.Name,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	return c.JSON(mapToWorkspaceDto(workspace))
}

func (h *Handler) DeleteWorkspace(c *fiber.Ctx) error {
	err := h.service.DeleteWorkspace(c.Context(), DeleteWorkspaceInput{
		ID: c.Params("id"),
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func mapToWorkspaceDto(workspace Workspace) dtos.WorkspaceDto {
	return dtos.WorkspaceDto{
		ID:        workspace.ID,
		Name:      workspace.Name,
		OwnerID:   workspace.OwnerID,
		CreatedAt: workspace.CreatedAt,
		UpdatedAt: workspace.UpdatedAt,
	}
}
