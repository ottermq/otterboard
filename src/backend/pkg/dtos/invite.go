package dtos

import "time"

type InviteDto struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedBy   string    `json:"created_by"`
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	UsedAt      time.Time `json:"used_at"`
}
