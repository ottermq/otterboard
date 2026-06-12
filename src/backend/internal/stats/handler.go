package stats

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *StatsService
}

func NewHandler(service *StatsService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetStats(c *fiber.Ctx) error {
	workspaceId := c.Params("workspaceId")
	assigneeId, ok := common.CurrentUserID(c)
	if !ok {
		return common.HandlerError(c, common.ErrForbidden)
	}
	stats, err := h.service.GetStats(c.Context(), GetStatsInput{
		WorkspaceID: workspaceId,
		AssigneeID:  assigneeId,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(mapToStatsDto(stats))
}

func mapToStatsDto(s Stats) dtos.StatsDto {
	return dtos.StatsDto{
		TotalProjects:   s.TotalProjects,
		TotalIssues:     s.TotalIssues,
		AssignedIssues:  s.AssignedIssues,
		CompletedIssues: s.CompletedIssues,
		OverdueIssues:   s.OverdueIssues,
	}
}
