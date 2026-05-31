package api_keys_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/api_keys"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/stretchr/testify/require"
)

type mockApiKeyStore struct {
	createApiKeyFn func(ctx context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error)
	listApiKeysFn  func(ctx context.Context, workspaceID pgtype.UUID) ([]db.ApiKey, error)
	revokeApiKeyFn func(ctx context.Context, arg db.RevokeApiKeyParams) (db.ApiKey, error)
}

func (m *mockApiKeyStore) CreateApiKey(ctx context.Context, arg db.CreateApiKeyParams) (db.ApiKey, error) {
	return m.createApiKeyFn(ctx, arg)
}

func (m *mockApiKeyStore) ListApiKeys(ctx context.Context, workspaceID pgtype.UUID) ([]db.ApiKey, error) {
	return m.listApiKeysFn(ctx, workspaceID)
}

func (m *mockApiKeyStore) RevokeApiKey(ctx context.Context, arg db.RevokeApiKeyParams) (db.ApiKey, error) {
	return m.revokeApiKeyFn(ctx, arg)
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

	service := api_keys.NewApiKeyService(store)
	apiKey, rawKey, err := service.CreateApiKey(context.Background(), api_keys.CreateApiKeyInput{
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

	service := api_keys.NewApiKeyService(store)
	_, rawKey, err := service.CreateApiKey(context.Background(), api_keys.CreateApiKeyInput{
		WorkspaceID: "not-a-uuid",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Name:        "deploy key",
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Empty(t, rawKey)

	_, rawKey, err = service.CreateApiKey(context.Background(), api_keys.CreateApiKeyInput{
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

	service := api_keys.NewApiKeyService(store)
	apiKey, rawKey, err := service.CreateApiKey(context.Background(), api_keys.CreateApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Name:        "deploy key",
	})

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, apiKey)
	require.Empty(t, rawKey)
}

func TestListApiKeys_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	firstID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	secondID := mustUUID(t, "44444444-4444-4444-4444-444444444444")
	firstCreatedAt := time.Now().Add(-time.Hour)
	secondCreatedAt := time.Now()

	listCalled := false
	store := &mockApiKeyStore{
		listApiKeysFn: func(_ context.Context, gotWorkspaceID pgtype.UUID) ([]db.ApiKey, error) {
			listCalled = true
			require.Equal(t, workspaceID, gotWorkspaceID)
			return []db.ApiKey{
				{
					ID:          firstID,
					WorkspaceID: workspaceID,
					UserID:      userID,
					Name:        "deploy key",
					CreatedAt:   pgtype.Timestamptz{Time: firstCreatedAt, Valid: true},
				},
				{
					ID:          secondID,
					WorkspaceID: workspaceID,
					UserID:      userID,
					Name:        "ci key",
					CreatedAt:   pgtype.Timestamptz{Time: secondCreatedAt, Valid: true},
				},
			}, nil
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKeys, err := service.ListApiKeys(context.Background(), workspaceID.String())

	require.NoError(t, err)
	require.True(t, listCalled)
	require.Equal(t, []api_keys.ApiKey{
		{
			ID:          firstID.String(),
			WorkspaceID: workspaceID.String(),
			UserID:      userID.String(),
			Name:        "deploy key",
			CreatedAt:   firstCreatedAt,
		},
		{
			ID:          secondID.String(),
			WorkspaceID: workspaceID.String(),
			UserID:      userID.String(),
			Name:        "ci key",
			CreatedAt:   secondCreatedAt,
		},
	}, apiKeys)
}

func TestListApiKeys_InvalidWorkspaceID(t *testing.T) {
	store := &mockApiKeyStore{
		listApiKeysFn: func(_ context.Context, _ pgtype.UUID) ([]db.ApiKey, error) {
			t.Fatal("ListApiKeys should not be called with an invalid workspace ID")
			return nil, nil
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKeys, err := service.ListApiKeys(context.Background(), "not-a-uuid")

	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Empty(t, apiKeys)
}

func TestListApiKeys_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockApiKeyStore{
		listApiKeysFn: func(_ context.Context, _ pgtype.UUID) ([]db.ApiKey, error) {
			return nil, expectedErr
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKeys, err := service.ListApiKeys(context.Background(), "11111111-1111-1111-1111-111111111111")

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, apiKeys)
}

func TestRevokeApiKey_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	keyID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	createdAt := time.Now()

	revokeCalled := false
	store := &mockApiKeyStore{
		revokeApiKeyFn: func(_ context.Context, arg db.RevokeApiKeyParams) (db.ApiKey, error) {
			revokeCalled = true
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, keyID, arg.ID)
			return db.ApiKey{
				ID:          keyID,
				WorkspaceID: workspaceID,
				UserID:      userID,
				Name:        "deploy key",
				CreatedAt:   pgtype.Timestamptz{Time: createdAt, Valid: true},
			}, nil
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKey, err := service.RevokeApiKey(context.Background(), api_keys.RevokeApiKeyInput{
		WorkspaceID: workspaceID.String(),
		KeyID:       keyID.String(),
	})

	require.NoError(t, err)
	require.True(t, revokeCalled)
	require.Equal(t, api_keys.ApiKey{
		ID:          keyID.String(),
		WorkspaceID: workspaceID.String(),
		UserID:      userID.String(),
		Name:        "deploy key",
		CreatedAt:   createdAt,
	}, apiKey)
}

func TestRevokeApiKey_InvalidIDs(t *testing.T) {
	store := &mockApiKeyStore{
		revokeApiKeyFn: func(_ context.Context, _ db.RevokeApiKeyParams) (db.ApiKey, error) {
			t.Fatal("RevokeApiKey should not be called with invalid IDs")
			return db.ApiKey{}, nil
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKey, err := service.RevokeApiKey(context.Background(), api_keys.RevokeApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		KeyID:       "not-a-uuid",
	})
	require.ErrorIs(t, err, api_keys.ErrInvalidApiKeyID)
	require.Empty(t, apiKey)

	apiKey, err = service.RevokeApiKey(context.Background(), api_keys.RevokeApiKeyInput{
		WorkspaceID: "not-a-uuid",
		KeyID:       "33333333-3333-3333-3333-333333333333",
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Empty(t, apiKey)
}

func TestRevokeApiKey_NotFound(t *testing.T) {
	store := &mockApiKeyStore{
		revokeApiKeyFn: func(_ context.Context, _ db.RevokeApiKeyParams) (db.ApiKey, error) {
			return db.ApiKey{}, pgx.ErrNoRows
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKey, err := service.RevokeApiKey(context.Background(), api_keys.RevokeApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		KeyID:       "33333333-3333-3333-3333-333333333333",
	})

	require.ErrorIs(t, err, api_keys.ErrApiKeyNotFound)
	require.Empty(t, apiKey)
}

func TestRevokeApiKey_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockApiKeyStore{
		revokeApiKeyFn: func(_ context.Context, _ db.RevokeApiKeyParams) (db.ApiKey, error) {
			return db.ApiKey{}, expectedErr
		},
	}

	service := api_keys.NewApiKeyService(store)
	apiKey, err := service.RevokeApiKey(context.Background(), api_keys.RevokeApiKeyInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		KeyID:       "33333333-3333-3333-3333-333333333333",
	})

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, apiKey)
}
