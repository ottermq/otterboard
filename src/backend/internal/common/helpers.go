package common

import "github.com/gofiber/fiber/v2"

const (
	WorkspaceRoleKey = "workspaceRole"
)

func CurrentUserID(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals("userID").(string)
	return userID, ok && userID != ""
}

func CurrentWorkspaceRole(c *fiber.Ctx) (string, bool) {
	role, ok := c.Locals(WorkspaceRoleKey).(string)
	return role, ok && role != ""
}
