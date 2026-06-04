package middleware

import (
	"context"
	"errors"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

type WorkspaceAuthStore interface {
	GetMember(ctx context.Context, arg db.GetMemberParams) (db.WorkspaceMember, error)
}

func RequireWorkspaceMember(store WorkspaceAuthStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := common.CurrentUserID(c)
		if !ok {
			return common.HandlerError(c, common.ErrUnauthorized)
		}
		var userUUID pgtype.UUID
		if err := userUUID.Scan(userID); err != nil {
			return common.HandlerError(c, common.ErrInvalidUserID)
		}
		var workspaceUUID pgtype.UUID
		if err := workspaceUUID.Scan(c.Params("workspaceId")); err != nil {
			return common.HandlerError(c, common.ErrInvalidWorkspaceID)
		}
		member, err := store.GetMember(c.Context(), db.GetMemberParams{
			WorkspaceID: workspaceUUID,
			UserID:      userUUID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return common.HandlerError(c, common.ErrForbidden)
		}
		if err != nil {
			return common.HandlerError(c, err)
		}
		c.Locals(common.WorkspaceRoleKey, member.Role)
		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(common.WorkspaceRoleKey).(string)
		if !ok {
			return common.HandlerError(c, common.ErrForbidden)
		}
		if !slices.Contains(allowedRoles, role) {
			return common.HandlerError(c, common.ErrForbidden)
		}
		return c.Next()
	}
}
