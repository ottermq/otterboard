package members

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *MemberService
}

func NewHandler(service *MemberService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) ListMembers(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	members, err := h.service.ListMembers(c.Context(), workspaceID)
	if err != nil {
		return common.HandlerError(c, err)
	}

	dtos := make([]dtos.WorkspaceMemberDto, 0, len(members))
	for _, m := range members {
		dtos = append(dtos, mapToMemberDto(m))
	}
	return c.JSON(dtos)
}

func (h *Handler) UpdateMemberRole(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	userID := c.Params("userId")

	var req struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	member, err := h.service.UpdateMemberRole(c.Context(), UpdateMemberRoleInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		NewRole:     req.Role,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	return c.JSON(mapToMemberDto(member))
}

func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	userID := c.Params("userId")

	err := h.service.RemoveMember(c.Context(), workspaceID, userID)
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapToMemberDto(m Member) dtos.WorkspaceMemberDto {
	return dtos.WorkspaceMemberDto{
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        m.Role,
		JoinedAt:    m.JoinedAt,
	}
}
