package api_keys

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/pkg/dtos"
)

type Handler struct {
	service *ApiKeyService
}

func NewHandler(service *ApiKeyService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateApiKey(c *fiber.Ctx) error {
	userID, ok := common.CurrentUserID(c)
	if !ok {
		return common.Unauthorized(c)
	}

	workspaceID := c.Params("workspaceId")

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": "invalid request"})
	}

	apiKey, rawKey, err := h.service.CreateApiKey(c.Context(), CreateApiKeyInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        req.Name,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(struct {
		dtos.ApiKeyDto
		Key string `json:"key"`
	}{
		ApiKeyDto: mapToApiKeyDto(apiKey),
		Key:       rawKey,
	})
}

func (h *Handler) ListApiKeys(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")

	apiKeys, err := h.service.ListApiKeys(c.Context(), workspaceID)
	if err != nil {
		return common.HandlerError(c, err)
	}
	response := make([]dtos.ApiKeyDto, len(apiKeys))
	for i, apiKey := range apiKeys {
		response[i] = mapToApiKeyDto(apiKey)
	}
	return c.JSON(response)
}

func (h *Handler) RevokeApiKey(c *fiber.Ctx) error {
	workspaceId := c.Params("workspaceId")
	keyId := c.Params("keyId")

	_, err := h.service.RevokeApiKey(c.Context(), RevokeApiKeyInput{
		WorkspaceID: workspaceId,
		KeyID:       keyId,
	})
	if err != nil {
		return common.HandlerError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func mapToApiKeyDto(apiKey ApiKey) dtos.ApiKeyDto {
	return dtos.ApiKeyDto{
		ID:          apiKey.ID,
		WorkspaceID: apiKey.WorkspaceID,
		UserID:      apiKey.UserID,
		Name:        apiKey.Name,
		CreatedAt:   apiKey.CreatedAt,
	}
}
