package projects

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *ProjectService
}

func NewHandler(service *ProjectService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateProject(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`
	}
	if err := c.BodyParser(&req); err != nil {
		return common.HandlerError(c, common.ErrBadRequest)
	}
	project, err := h.service.CreateProject(c.Context(), CreateProjectInput{
		WorkspaceID: c.Params("workspaceId"),
		Name:        req.Name,
		ImageUrl:    req.ImageUrl,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(mapToProjectDto(project))
}

func (h *Handler) GetProject(c *fiber.Ctx) error {
	project, err := h.service.GetProjectByID(c.Context(), GetProjectByIdInput{
		ID:          c.Params("id"),
		WorkspaceID: c.Params("workspaceId"),
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToProjectDto(project))
}

func (h *Handler) ListProjects(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	page := max(int32(c.QueryInt("page")), 1)
	limit := int32(c.QueryInt("limit"))
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	total, err := h.service.CountProjectsByWorkspace(c.Context(), workspaceID)
	if err != nil {
		return common.HandlerError(c, err)
	}

	projectList, err := h.service.ListProjectsByWorkspace(c.Context(), ListProjectsByWorkspaceInput{
		WorkspaceID: workspaceID,
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapListToPaginatedResponse(projectList, total, page, limit))
}

func (h *Handler) UpdateProject(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		ImageUrl string `json:"image_url"`
	}
	if err := c.BodyParser(&req); err != nil {
		return common.HandlerError(c, common.ErrBadRequest)
	}
	input := UpdateProjectInput{
		ID:          c.Params("id"),
		WorkspaceID: c.Params("workspaceId"),
		Name:        req.Name,
		ImageUrl:    req.ImageUrl,
	}
	project, err := h.service.UpdateProject(c.Context(), input)
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToProjectDto(project))
}

func (h *Handler) DeleteProject(c *fiber.Ctx) error {
	input := DeleteProjectInput{
		WorkspaceID: c.Params("workspaceId"),
		ID:          c.Params("id"),
	}
	err := h.service.DeleteProject(c.Context(), input)
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapToProjectDto(project Project) dtos.ProjectDto {
	return dtos.ProjectDto{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		Name:        project.Name,
		ImageURL:    project.ImageUrl,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func mapListToPaginatedResponse(projectList []Project, total int64, page, limit int32) dtos.PaginatedResponse[dtos.ProjectDto] {
	dtoList := make([]dtos.ProjectDto, len(projectList))
	for idx, project := range projectList {
		dtoList[idx] = mapToProjectDto(project)
	}
	return dtos.PaginatedResponse[dtos.ProjectDto]{
		Data:  dtoList,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}
