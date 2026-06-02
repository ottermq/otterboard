package projects_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/projects"
	"github.com/stretchr/testify/require"
)

type mockProjectStore struct {
	createProjectFn  func(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
	getProjectByIDFn func(ctx context.Context, arg db.GetProjectByIDParams) (db.Project, error)
}

func (m *mockProjectStore) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	if m.createProjectFn == nil {
		panic("unexpected call to CreateProject")
	}
	return m.createProjectFn(ctx, arg)
}

func (m *mockProjectStore) GetProjectByID(ctx context.Context, arg db.GetProjectByIDParams) (db.Project, error) {
	if m.getProjectByIDFn == nil {
		panic("unexpected call to GetProjectByID")
	}
	return m.getProjectByIDFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateProject_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			return db.Project{
				ID:          projectID,
				WorkspaceID: arg.WorkspaceID,
				Name:        arg.Name,
				ImageUrl:    arg.ImageUrl,
				CreatedAt:   pgtype.Timestamptz{},
				UpdatedAt:   pgtype.Timestamptz{},
			}, nil
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: workspaceID.String(),
		Name:        "Test Project",
		ImageUrl:    "test_image_url",
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), project.WorkspaceID)
	require.Equal(t, "Test Project", project.Name)
}

func TestCreateProject_ValidationErrors(t *testing.T) {
	validWorkspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name      string
		input     projects.CreateProjectInput
		wantError error
	}{
		{
			name: "invalid workspace id",
			input: projects.CreateProjectInput{
				WorkspaceID: "invalid UUID",
				Name:        "Test Project",
			},
			wantError: common.ErrInvalidWorkspaceID,
		},
		{
			name: "empty name",
			input: projects.CreateProjectInput{
				WorkspaceID: validWorkspaceID.String(),
				Name:        "",
			},
			wantError: projects.ErrInvalidProjectName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockProjectStore{
				createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
					t.Fatal("CreateProject should not be called with invalid input")
					return db.Project{}, nil
				},
			}

			service := projects.NewProjectService(store)
			project, err := service.CreateProject(context.Background(), tt.input)

			require.ErrorIs(t, err, tt.wantError)
			require.Equal(t, projects.Project{}, project)
		})
	}
}

func TestCreateProject_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := errors.New("generic storage error")

	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			return db.Project{}, expectedErr
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: workspaceID.String(),
		Name:        "Generic Name",
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, projects.Project{}, project)
}

func TestGetProjectByID_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockProjectStore{
		getProjectByIDFn: func(ctx context.Context, arg db.GetProjectByIDParams) (db.Project, error) {
			return db.Project{
				ID:          projectID,
				WorkspaceID: workspaceID,
			}, nil
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.GetProjectByID(context.Background(), projects.GetProjectByIdInput{
		ID:          projectID.String(),
		WorkspaceID: workspaceID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, projectID.String(), project.ID)
	require.Equal(t, workspaceID.String(), project.WorkspaceID)
}

func TestGetProjectByID_ValidationErrors(t *testing.T) {
	validWorkspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validProjectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	tests := []struct {
		name      string
		input     projects.GetProjectByIdInput
		wantError error
	}{
		{
			name: "invalid workspace id",
			input: projects.GetProjectByIdInput{
				WorkspaceID: "invalid UUID",
				ID:          validProjectID.String(),
			},
			wantError: common.ErrInvalidWorkspaceID,
		},
		{
			name: "invalid project id",
			input: projects.GetProjectByIdInput{
				WorkspaceID: validWorkspaceID.String(),
				ID:          "invalid UUID",
			},
			wantError: projects.ErrInvalidProjectID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockProjectStore{
				getProjectByIDFn: func(ctx context.Context, arg db.GetProjectByIDParams) (db.Project, error) {
					t.Fatal("GetProjectByID should not be called with invalid input")
					return db.Project{}, nil
				},
			}

			service := projects.NewProjectService(store)
			project, err := service.GetProjectByID(context.Background(), tt.input)

			require.ErrorIs(t, err, tt.wantError)
			require.Equal(t, projects.Project{}, project)
		})
	}
}

func TestGetProjectByID_ProjectNotFound(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := projects.ErrProjectNotFound

	store := &mockProjectStore{
		getProjectByIDFn: func(_ context.Context, arg db.GetProjectByIDParams) (db.Project, error) {
			return db.Project{}, pgx.ErrNoRows
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.GetProjectByID(context.Background(), projects.GetProjectByIdInput{
		WorkspaceID: workspaceID.String(),
		ID:          workspaceID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, projects.Project{}, project)
}

func TestGetProjectByID_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := errors.New("generic storage error")

	store := &mockProjectStore{
		getProjectByIDFn: func(_ context.Context, arg db.GetProjectByIDParams) (db.Project, error) {
			return db.Project{}, expectedErr
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.GetProjectByID(context.Background(), projects.GetProjectByIdInput{
		WorkspaceID: workspaceID.String(),
		ID:          workspaceID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, projects.Project{}, project)
}
