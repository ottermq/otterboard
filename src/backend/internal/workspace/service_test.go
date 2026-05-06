package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/workspace"
	"github.com/stretchr/testify/require"
)

type mockWorkspaceStore struct {
	createWorkspaceFn        func(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error)
	getWorkspaceByIDFn       func(ctx context.Context, id pgtype.UUID) (db.Workspace, error)
	getWorkspacesByOwnerIDFn func(ctx context.Context, ownerID pgtype.UUID) ([]db.Workspace, error)
	updateWorkspaceFn        func(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error)
	deleteWorkspaceFn        func(ctx context.Context, id pgtype.UUID) error
}

func (m *mockWorkspaceStore) CreateWorkspace(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
	return m.createWorkspaceFn(ctx, arg)
}

func (m *mockWorkspaceStore) GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (db.Workspace, error) {
	return m.getWorkspaceByIDFn(ctx, id)
}

func (m *mockWorkspaceStore) GetWorkspacesByOwnerID(ctx context.Context, ownerID pgtype.UUID) ([]db.Workspace, error) {
	return m.getWorkspacesByOwnerIDFn(ctx, ownerID)
}

func (m *mockWorkspaceStore) UpdateWorkspace(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
	return m.updateWorkspaceFn(ctx, arg)
}

func (m *mockWorkspaceStore) DeleteWorkspace(ctx context.Context, id pgtype.UUID) error {
	return m.deleteWorkspaceFn(ctx, id)
}

func TestCreateWorkspace_Success(t *testing.T) {
	ownerID := pgtype.UUID{}
	err := ownerID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		createWorkspaceFn: func(_ context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				Name:    arg.Name,
				OwnerID: arg.OwnerID,
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	workspace, err := service.CreateWorkspace(context.Background(), workspace.CreateWorkspaceInput{
		Name:    "Test Workspace",
		OwnerID: ownerID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, "Test Workspace", workspace.Name)
	require.Equal(t, ownerID.String(), workspace.OwnerID)
}

func TestGetWorkspaceByID_Success(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	ws, err := service.GetWorkspaceByID(context.Background(), workspaceID.String())
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), ws.ID)
	require.Equal(t, "Test Workspace", ws.Name)
	require.Equal(t, ownerID.String(), ws.OwnerID)
}

func TestGetWorkspaceByID_NotFound(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{}, pgx.ErrNoRows
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err = service.GetWorkspaceByID(context.Background(), workspaceID.String())
	require.ErrorIs(t, err, workspace.ErrWorkspaceNotFound)
}

func TestCreateWorkspace_InvalidOwnerID(t *testing.T) {
	store := &mockWorkspaceStore{
		createWorkspaceFn: func(_ context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				Name:    arg.Name,
				OwnerID: pgtype.UUID{},
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err := service.CreateWorkspace(context.Background(), workspace.CreateWorkspaceInput{
		Name:    "Test Workspace",
		OwnerID: "invalid-uuid",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, workspace.ErrInvalidOwnerID)
}

func TestGetWorkspacesByOwnerID_Success(t *testing.T) {
	ownerID := pgtype.UUID{}
	err := ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspacesByOwnerIDFn: func(_ context.Context, id pgtype.UUID) ([]db.Workspace, error) {
			return []db.Workspace{
				{
					ID:      pgtype.UUID{Bytes: [16]byte{0x89, 0xe4, 0x10, 0x58, 0x2e, 0x01, 0x48, 0x7e, 0xad, 0x8c, 0x8e, 0x9e, 0x35, 0xc3, 0x19, 0x87}, Valid: true},
					Name:    "Workspace 1",
					OwnerID: id,
				},
				{
					ID:      pgtype.UUID{Bytes: [16]byte{0x98, 0xe4, 0x10, 0x58, 0x2e, 0x01, 0x48, 0x7e, 0xad, 0x8c, 0x8e, 0x9e, 0x35, 0xc3, 0x19, 0x88}, Valid: true},
					Name:    "Workspace 2",
					OwnerID: id,
				},
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	workspaces, err := service.GetWorkspacesByOwnerID(context.Background(), ownerID.String())
	require.NoError(t, err)
	require.Len(t, workspaces, 2)
	require.Equal(t, "Workspace 1", workspaces[0].Name)
	require.Equal(t, "Workspace 2", workspaces[1].Name)
	require.Equal(t, ownerID.String(), workspaces[0].OwnerID)
	require.Equal(t, ownerID.String(), workspaces[1].OwnerID)
}

func TestGetWorkspacesByOwnerID_InvalidOwnerID(t *testing.T) {
	store := &mockWorkspaceStore{
		getWorkspacesByOwnerIDFn: func(_ context.Context, id pgtype.UUID) ([]db.Workspace, error) {
			return []db.Workspace{}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err := service.GetWorkspacesByOwnerID(context.Background(), "invalid-uuid")
	require.Error(t, err)
}

func TestUpdateWorkspace_Success(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				ID:   arg.ID,
				Name: arg.Name,
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	updatedWorkspace, err := service.UpdateWorkspace(context.Background(), workspace.UpdateWorkspaceInput{
		ID:          workspaceID.String(),
		Name:        "Updated Workspace Name",
		RequestorID: ownerID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), updatedWorkspace.ID)
	require.Equal(t, "Updated Workspace Name", updatedWorkspace.Name)
}

func TestUpdateWorkspace_InvalidWorkspaceID(t *testing.T) {
	store := &mockWorkspaceStore{
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err := service.UpdateWorkspace(context.Background(), workspace.UpdateWorkspaceInput{
		ID:          "invalid-uuid",
		Name:        "Updated Workspace Name",
		RequestorID: "123e4567-e89b-12d3-a456-426614174000",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
}

func TestUpdateWorkspace_WorkspaceNotFound(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, _ pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{}, pgx.ErrNoRows
		},
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, pgx.ErrNoRows
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err = service.UpdateWorkspace(context.Background(), workspace.UpdateWorkspaceInput{
		ID:          workspaceID.String(),
		Name:        "Updated Workspace Name",
		RequestorID: "123e4567-e89b-12d3-a456-426614174000",
	})
	require.ErrorIs(t, err, workspace.ErrWorkspaceNotFound)
}

func TestUpdateWorkspace_Forbidden(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err = service.UpdateWorkspace(context.Background(), workspace.UpdateWorkspaceInput{
		ID:          workspaceID.String(),
		Name:        "Updated Workspace Name",
		RequestorID: "123e4567-e89b-12d3-a456-426614174001", // different owner ID than existing workspace
	})
	require.ErrorIs(t, err, common.ErrForbidden)
}

func TestUpdateWorkspace_InvalidRequestorID(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	_, err = service.UpdateWorkspace(context.Background(), workspace.UpdateWorkspaceInput{
		ID:          workspaceID.String(),
		Name:        "Updated Workspace Name",
		RequestorID: "invalid-uuid",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidRequestorID)
}

func TestDeleteWorkspace_Success(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	deleteCalled := false
	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			require.Equal(t, workspaceID, id)
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
		deleteWorkspaceFn: func(_ context.Context, id pgtype.UUID) error {
			require.Equal(t, workspaceID, id)
			deleteCalled = true
			return nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: ownerID.String(),
	})
	require.NoError(t, err)
	require.True(t, deleteCalled)
}

func TestDeleteWorkspace_InvalidWorkspaceID(t *testing.T) {
	store := &mockWorkspaceStore{}

	service := workspace.NewWorkspaceService(store)
	err := service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          "invalid-uuid",
		RequestorID: "123e4567-e89b-12d3-a456-426614174000",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
}

func TestDeleteWorkspace_InvalidRequestorID(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: "invalid-uuid",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidRequestorID)
}

func TestDeleteWorkspace_WorkspaceNotFound(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			require.Equal(t, workspaceID, id)
			return db.Workspace{}, pgx.ErrNoRows
		},
	}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: "123e4567-e89b-12d3-a456-426614174000",
	})
	require.ErrorIs(t, err, workspace.ErrWorkspaceNotFound)
}

func TestDeleteWorkspace_Forbidden(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
	}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: "123e4567-e89b-12d3-a456-426614174001",
	})
	require.ErrorIs(t, err, common.ErrForbidden)
}

func TestDeleteWorkspace_GetWorkspaceError(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)
	expectedErr := errors.New("get workspace failed")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			require.Equal(t, workspaceID, id)
			return db.Workspace{}, expectedErr
		},
	}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: ownerID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
}

func TestDeleteWorkspace_DeleteError(t *testing.T) {
	workspaceID := pgtype.UUID{}
	err := workspaceID.Scan("89e41058-2e01-487e-ad8c-8e9e35c31987")
	require.NoError(t, err)
	ownerID := pgtype.UUID{}
	err = ownerID.Scan("123e4567-e89b-12d3-a456-426614174000")
	require.NoError(t, err)
	expectedErr := errors.New("delete workspace failed")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:      id,
				Name:    "Test Workspace",
				OwnerID: ownerID,
			}, nil
		},
		deleteWorkspaceFn: func(_ context.Context, id pgtype.UUID) error {
			require.Equal(t, workspaceID, id)
			return expectedErr
		},
	}

	service := workspace.NewWorkspaceService(store)
	err = service.DeleteWorkspace(context.Background(), workspace.DeleteWorkspaceInput{
		ID:          workspaceID.String(),
		RequestorID: ownerID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
}
