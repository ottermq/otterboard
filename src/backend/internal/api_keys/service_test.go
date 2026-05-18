package apikeys_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	apikeys "github.com/ottermq/otterboard/src/backend/internal/api_keys"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/stretchr/testify/require"
)

type mockApiKeyStore struct {
	createApiKeyFn func(ctx context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error)
}

func (m *mockApiKeyStore) CreateApiKey(ctx context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error) {
	return m.createApiKeyFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateApiKey_Success(t *testing.T) {
	id := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	createdAt := time.Now()

	createCalled := false
	var storedHash string
	store := &mockApiKeyStore{
		createApiKeyFn: func(_ context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error) {
			createCalled = true
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, userID, arg.UserID)
			require.Equal(t, "deploy key", arg.Name)
			require.NotEmpty(t, arg.KeyHash)
			storedHash = arg.KeyHash

			decodedHash, err := base64.RawURLEncoding.DecodeString(arg.KeyHash)
			require.NoError(t, err)
			require.Len(t, decodedHash, sha256.Size)

			return db.ApiKey{
				ID:          id,
				WorkspaceID: arg.WorkspaceID,
				UserID:      arg.UserID,
				Name:        arg.Name,
				KeyHash:     arg.KeyHash,
				CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
			}, nil
		},
	}

	service := apikeys.NewApiKeyService(store)
	apiKey, rawKey, err := service.CreateApiKey(context.Background(), apikeys.CreateApiKeyInput{
		WorkspaceID: workspaceID.String(),
		UserID:      userID.String(),
		Name:        "deploy key",
	})

	require.NoError(t, err)
	require.True(t, createCalled)
	require.Equal(t, id.String(), apiKey.ID)
	require.Equal(t, workspaceID.String(), apiKey.WorkspaceID)
	require.Equal(t, userID.String(), apiKey.UserID)
	require.Equal(t, "deploy key", apiKey.Name)
	require.Equal(t, createdAt, apiKey.CreatedAt)
	require.NotEmpty(t, rawKey)
	require.Equal(t, "ob_", rawKey[:len("ob_")])

	decodedRawKey, err := base64.RawURLEncoding.DecodeString(rawKey[len("ob_"):])
	require.NoError(t, err)
	require.Len(t, decodedRawKey, 32)

	hash := sha256.Sum256([]byte(rawKey))
	require.Equal(t, base64.RawURLEncoding.EncodeToString(hash[:]), storedHash)
}

func TestCreateApiKey_InvalidIDs(t *testing.T) {
	store := &mockApiKeyStore{
		createApiKeyFn: func(_ context.Context, _ db.CreateApiKeyParams) (db.ApiKey, error) {
			t.Fatal("CreateApiKey should not be called with invalid IDs")
			return db.ApiKey{}, nil
		},
	}

	service := apikeys.NewApiKeyService(store)
	_, rawKey, err := service.CreateApiKey(context.Background(), apikeys.CreateApiKeyInput{
		WorkspaceID: "not-a-uuid",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Name:        "deploy key",
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Empty(t, rawKey)

	_, rawKey, err = service.CreateApiKey(context.Background(), apikeys.CreateApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "not-a-uuid",
		Name:        "deploy key",
	})
	require.ErrorIs(t, err, common.ErrInvalidUserID)
	require.Empty(t, rawKey)
}

func TestCreateApiKey_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockApiKeyStore{
		createApiKeyFn: func(_ context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error) {
			require.Equal(t, "deploy key", arg.Name)
			require.NotEmpty(t, arg.KeyHash)
			return db.ApiKey{}, expectedErr
		},
	}

	service := apikeys.NewApiKeyService(store)
	apiKey, rawKey, err := service.CreateApiKey(context.Background(), apikeys.CreateApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Name:        "deploy key",
	})

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, apiKey)
	require.Empty(t, rawKey)
}
