package api_keys

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrInvalidApiKeyID = common.NewAppError(400, "invalid API key ID")
	ErrApiKeyNotFound  = common.NewAppError(404, "API key not found")
)

type ApiKeyStore interface {
	CreateApiKey(ctx context.Context, input db.CreateApiKeyParams) (db.ApiKey, error)
	ListApiKeys(ctx context.Context, workspaceID pgtype.UUID) ([]db.ApiKey, error)
	RevokeApiKey(ctx context.Context, arg db.RevokeApiKeyParams) (db.ApiKey, error)
}

type ApiKeyService struct {
	store ApiKeyStore
}

type CreateApiKeyInput struct {
	WorkspaceID string
	UserID      string
	Name        string
}

type RevokeApiKeyInput struct {
	WorkspaceID string
	KeyID       string
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

func (s *ApiKeyService) ListApiKeys(ctx context.Context, workspaceID string) ([]ApiKey, error) {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(workspaceID); err != nil {
		return nil, common.ErrInvalidWorkspaceID
	}

	keys, err := s.store.ListApiKeys(ctx, workspaceUUID)
	if err != nil {
		return nil, err
	}

	result := make([]ApiKey, 0, len(keys))
	for _, k := range keys {
		result = append(result, mapDbApiKeyToDomain(k))
	}
	return result, nil
}

func (s *ApiKeyService) RevokeApiKey(ctx context.Context, input RevokeApiKeyInput) (ApiKey, error) {
	var apiKeyID pgtype.UUID
	if err := apiKeyID.Scan(input.KeyID); err != nil {
		return ApiKey{}, ErrInvalidApiKeyID
	}

	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return ApiKey{}, common.ErrInvalidWorkspaceID
	}

	dbApikey, err := s.store.RevokeApiKey(ctx, db.RevokeApiKeyParams{
		WorkspaceID: workspaceID,
		ID:          apiKeyID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ApiKey{}, ErrApiKeyNotFound
	}
	if err != nil {
		return ApiKey{}, err
	}
	return mapDbApiKeyToDomain(dbApikey), nil
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
