package apikeys

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

type ApiKeyStore interface {
	CreateApiKey(ctx context.Context, input db.CreateApiKeyParams) (db.ApiKey, error)
}

type ApiKeyService struct {
	store ApiKeyStore
}

type CreateApiKeyInput struct {
	WorkspaceID string
	UserID      string
	Name        string
}

func NewApiKeyService(store ApiKeyStore) *ApiKeyService {
	return &ApiKeyService{
		store: store,
	}
}

type ApiKey struct {
	ID          string
	WorkspaceID string
	UserID      string
	Name        string
	CreatedAt   time.Time
}

func (s *ApiKeyService) CreateApiKey(ctx context.Context, input CreateApiKeyInput) (ApiKey, string, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return ApiKey{}, "", common.ErrInvalidWorkspaceID
	}

	var userID pgtype.UUID
	if err := userID.Scan(input.UserID); err != nil {
		return ApiKey{}, "", common.ErrInvalidUserID
	}

	rawKey, err := generateRawApiKey()
	if err != nil {
		return ApiKey{}, "", err
	}

	hash := sha256.Sum256([]byte(rawKey))
	encodedHash := base64.RawURLEncoding.EncodeToString(hash[:])

	dbInput := db.CreateApiKeyParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        input.Name,
		KeyHash:     encodedHash,
	}
	apiKey, err := s.store.CreateApiKey(ctx, dbInput)
	if err != nil {
		return ApiKey{}, "", err
	}
	return mapDbApiKeyToDomain(apiKey), rawKey, nil
}

func generateRawApiKey() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return "ob_" + base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func mapDbApiKeyToDomain(dbKey db.ApiKey) ApiKey {
	return ApiKey{
		ID:          dbKey.ID.String(),
		WorkspaceID: dbKey.WorkspaceID.String(),
		UserID:      dbKey.UserID.String(),
		Name:        dbKey.Name,
		CreatedAt:   dbKey.CreatedAt.Time,
	}
}
