package invites

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
)

const (
	DefaultInviteExpiration = 7 * 24 * time.Hour
)

type Handler struct {
	service *InviteService
}

func NewHandler(service *InviteService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GenerateInvite(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.Unauthorized(c)
	}
	invite, err := h.service.GenerateInvite(c.Context(), GenerateInviteInput{
		WorkspaceID: c.Params("workspaceId"),
		CreatedBy:   userID,
		ExpiresIn:   DefaultInviteExpiration,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(invite)
}

func (h *Handler) GetInvite(c *fiber.Ctx) error {
	invite, err := h.service.GetInvite(c.Context(), c.Params("token"))
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(invite)
}

func (h *Handler) AcceptInvite(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.Unauthorized(c)
	}
	err := h.service.AcceptInvite(c.Context(), AcceptInviteInput{
		Token:  c.Params("token"),
		UserID: userID,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.JSON(fiber.Map{"message": "invite accepted"})
}
