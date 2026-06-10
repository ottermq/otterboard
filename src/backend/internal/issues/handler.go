package issues

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *IssueService
}

func NewHandler(service *IssueService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateIssue(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.HandlerError(c, common.ErrUnauthorized)
	}
	var req struct {
		Title      string     `json:"title"`
		Overview   string     `json:"overview"`
		Type       string     `json:"type"`
		AssigneeID string     `json:"assignee_id"`
		DueDate    *time.Time `json:"due_date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return common.HandlerError(c, common.ErrBadRequest)
	}
	issue, err := h.service.CreateIssue(c.Context(), CreateIssueInput{
		ProjectID:  c.Params("projectId"),
		Title:      req.Title,
		Overview:   req.Overview,
		Type:       req.Type,
		AssigneeID: req.AssigneeID,
		CreatedBy:  userID,
		DueDate:    req.DueDate,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(mapToIssueDto(issue))
}

func (h *Handler) ListIssues(c *fiber.Ctx) error {
	projectId := c.Params("projectId")
	page := max(int32(c.QueryInt("page")), 1)
	limit := int32(c.QueryInt("limit"))
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	filters, err := resolveFilters(c)
	if err != nil {
		return common.HandlerError(c, err)
	}
	total, err := h.service.CountIssuesByProjectFiltered(c.Context(), CountIssuesByProjectFilteredInput{
		ProjectID: projectId,
		Filters:   filters,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	issueList, err := h.service.ListIssuesByProjectFiltered(c.Context(), ListIssuesByProjectInput{
		ProjectID: projectId,
		Filters:   filters,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapListToPaginatedResponse(issueList, total, page, limit))
}

func (h *Handler) ListIssuesByWorkspace(c *fiber.Ctx) error {
	workspaceId := c.Params("workspaceId")
	projectId := c.Query("project")
	page := max(int32(c.QueryInt("page")), 1)
	limit := int32(c.QueryInt("limit"))
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	filters, err := resolveFilters(c)
	if err != nil {
		return common.HandlerError(c, err)
	}
	total, err := h.service.CountIssuesByWorkspaceFiltered(c.Context(), CountIssuesByWorkspaceFilteredInput{
		WorkspaceID:     workspaceId,
		ProjectIDFilter: projectId,
		Filters:         filters,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	issueList, err := h.service.ListIssuesByWorkspaceFiltered(c.Context(), ListIssuesByWorkspaceInput{
		WorkspaceID:     workspaceId,
		ProjectIDFilter: projectId,
		Filters:         filters,
		Page:            page,
		Limit:           limit,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapListToPaginatedResponse(issueList, total, page, limit))
}

func (h *Handler) GetIssue(c *fiber.Ctx) error {
	issue, err := h.service.GetIssueByID(c.Context(), GetIssueByIdInput{
		ID:        c.Params("id"),
		ProjectID: c.Params("projectId"),
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToIssueDto(issue))
}

func (h *Handler) UpdateIssue(c *fiber.Ctx) error {
	var req struct {
		Title      string     `json:"title"`
		Overview   string     `json:"overview"`
		Type       string     `json:"type"`
		Status     string     `json:"status"`
		Position   float64    `json:"position"`
		AssigneeID string     `json:"assignee_id"`
		DueDate    *time.Time `json:"due_date"`
	}
	if err := c.BodyParser(&req); err != nil {
		return common.HandlerError(c, common.ErrBadRequest)
	}
	issue, err := h.service.UpdateIssue(c.Context(), UpdateIssueInput{
		ID:         c.Params("id"),
		ProjectID:  c.Params("projectId"),
		Title:      req.Title,
		Overview:   req.Overview,
		Type:       req.Type,
		Status:     req.Status,
		Position:   req.Position,
		AssigneeID: req.AssigneeID,
		DueDate:    req.DueDate,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToIssueDto(issue))
}

func (h *Handler) DeleteIssue(c *fiber.Ctx) error {
	err := h.service.DeleteIssue(c.Context(), DeleteIssueInput{
		ID:        c.Params("id"),
		ProjectID: c.Params("projectId"),
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapToIssueDto(issue Issue) dtos.IssueDto {
	return dtos.IssueDto{
		ID:         issue.ID,
		ProjectID:  issue.ProjectID,
		Title:      issue.Title,
		Overview:   issue.Overview,
		Type:       issue.Type,
		Status:     issue.Status,
		CreatedBy:  issue.CreatedBy,
		AssigneeID: issue.AssigneeID,
		DueDate:    issue.DueDate,
		CreatedAt:  issue.CreatedAt,
		UpdatedAt:  issue.UpdatedAt,
	}
}

func mapListToPaginatedResponse(issueList []Issue, total int64, page, limit int32) dtos.PaginatedResponse[dtos.IssueDto] {
	dtoList := make([]dtos.IssueDto, len(issueList))
	for idx, issue := range issueList {
		dtoList[idx] = mapToIssueDto(issue)
	}
	return dtos.PaginatedResponse[dtos.IssueDto]{
		Data:  dtoList,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func resolveFilters(c *fiber.Ctx) (CommonFiltersInput, error) {
	dueBefore, err := parseDueTime(c, "due_before")
	if err != nil {
		return CommonFiltersInput{}, err
	}
	dueAfter, err := parseDueTime(c, "due_after")
	if err != nil {
		return CommonFiltersInput{}, err
	}

	assignee := c.Query("assignee")
	if assignee == "me" {
		assignee, _ = common.CurrentUserID(c)
	}

	return CommonFiltersInput{
		Status:     c.Query("status"),
		Type:       c.Query("type"),
		AssigneeID: assignee,
		DueBefore:  dueBefore,
		DueAfter:   dueAfter,
		SortBy:     c.Query("sort"),
		SortOrder:  c.Query("order"),
	}, nil
}

func parseDueTime(c *fiber.Ctx, query string) (*time.Time, error) {
	var dueTime *time.Time
	if dueQuery := c.Query(query); dueQuery != "" {
		parsedDueBefore, err := time.Parse("2006-01-02", dueQuery)
		if err != nil {
			return nil, common.ErrBadRequest
		}
		dueTime = &parsedDueBefore
	}
	return dueTime, nil
}
