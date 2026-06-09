package issues_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/issues"
	"github.com/stretchr/testify/require"
)

type mockIssueStore struct {
	getMaxPositionByProjectAndStatusFn func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error)
	createIssueFn                      func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error)
	getIssueByIDFn                     func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error)
	listIssuesByProjectFn              func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error)
	countIssuesByProjectFn             func(ctx context.Context, projectID pgtype.UUID) (int64, error)
	listIssuesByWorkspaceFn            func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error)
	countIssuesByWorkspaceFn           func(ctx context.Context, workspaceID pgtype.UUID) (int64, error)
	updateIssueFn                      func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error)
	deleteIssueFn                      func(ctx context.Context, arg db.DeleteIssueParams) error
}

func (m *mockIssueStore) GetMaxPositionByProjectAndStatus(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
	if m.getMaxPositionByProjectAndStatusFn == nil {
		panic("unexpected call to GetMaxPositionByProjectAndStatus")
	}
	return m.getMaxPositionByProjectAndStatusFn(ctx, arg)
}

func (m *mockIssueStore) CreateIssue(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
	if m.createIssueFn == nil {
		panic("unexpected call to CreateIssue")
	}
	return m.createIssueFn(ctx, arg)
}

func (m *mockIssueStore) GetIssueByID(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
	if m.getIssueByIDFn == nil {
		panic("unexpected call to GetIssueByID")
	}
	return m.getIssueByIDFn(ctx, arg)
}

func (m *mockIssueStore) ListIssuesByProject(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
	if m.listIssuesByProjectFn == nil {
		panic("unexpected call to ListIssuesByProject")
	}
	return m.listIssuesByProjectFn(ctx, arg)
}

func (m *mockIssueStore) CountIssuesByProject(ctx context.Context, projectID pgtype.UUID) (int64, error) {
	if m.countIssuesByProjectFn == nil {
		panic("unexpected call to CountIssuesByProject")
	}
	return m.countIssuesByProjectFn(ctx, projectID)
}

func (m *mockIssueStore) ListIssuesByWorkspace(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
	if m.listIssuesByWorkspaceFn == nil {
		panic("unexpected call to ListIssuesByWorkspace")
	}
	return m.listIssuesByWorkspaceFn(ctx, arg)
}

func (m *mockIssueStore) CountIssuesByWorkspace(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
	if m.countIssuesByWorkspaceFn == nil {
		panic("unexpected call to CountIssuesByWorkspace")
	}
	return m.countIssuesByWorkspaceFn(ctx, workspaceID)
}

func (m *mockIssueStore) UpdateIssue(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
	if m.updateIssueFn == nil {
		panic("unexpected call to UpdateIssue")
	}
	return m.updateIssueFn(ctx, arg)
}

func (m *mockIssueStore) DeleteIssue(ctx context.Context, arg db.DeleteIssueParams) error {
	if m.deleteIssueFn == nil {
		panic("unexpected call to DeleteIssue")
	}
	return m.deleteIssueFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateIssue_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	store := &mockIssueStore{
		getMaxPositionByProjectAndStatusFn: func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
			return float64(0.0), nil
		},
		createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
			return db.Issue{
				ID:         issueID,
				ProjectID:  arg.ProjectID,
				Title:      arg.Title,
				Overview:   arg.Overview,
				Type:       arg.Type,
				Status:     arg.Status,
				Position:   arg.Position,
				AssigneeID: arg.AssigneeID,
				CreatedBy:  arg.CreatedBy,
				DueDate:    arg.DueDate,
				CreatedAt:  pgtype.Timestamptz{},
				UpdatedAt:  pgtype.Timestamptz{},
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	issue, err := service.CreateIssue(context.Background(), issues.CreateIssueInput{
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		AssigneeID: userID.String(),
		CreatedBy:  userID.String(),
		DueDate:    nil,
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), issue.ID)
}

func TestCreateIssue_ValidationErrors(t *testing.T) {
	validProjectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validUserID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	tests := []struct {
		name      string
		input     issues.CreateIssueInput
		wantError error
	}{
		{
			name: "invalid project ID",
			input: issues.CreateIssueInput{
				ProjectID:  "invalid UUID",
				Title:      "title",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: common.ErrInvalidProjectID,
		},
		{
			name: "invalid assignee Id",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				AssigneeID: "invalid assignee",
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidAssigneeID,
		},
		{
			name: "invalid user Id",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  "invalid user ID",
				DueDate:    &time.Time{},
			},
			wantError: common.ErrInvalidUserID,
		},
		{
			name: "invalid title",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidTitle,
		},
		{
			name: "invalid issue type",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
					t.Fatal("CreateIssue should not be called with invalid input")
					return db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.CreateIssue(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantError)
			require.Equal(t, issues.Issue{}, got)
		})
	}
}

func TestCreateIssue_StoreError(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	expectedErr := errors.New("generic storage error")
	store := &mockIssueStore{
		getMaxPositionByProjectAndStatusFn: func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
			return float64(0.0), nil
		},
		createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
			return db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	issue, err := service.CreateIssue(context.Background(), issues.CreateIssueInput{
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		AssigneeID: userID.String(),
		CreatedBy:  userID.String(),
		DueDate:    nil,
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, issues.Issue{}, issue)
}

func TestGetIssueById_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{
				ID:        issueID,
				ProjectID: projectID,
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), got.ID)
	require.Equal(t, projectID.String(), got.ProjectID)
}

func TestGetIssueById_ValidationErrors(t *testing.T) {
	validProjectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validIssueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	tests := []struct {
		name    string
		input   issues.GetIssueByIdInput
		wantErr error
	}{
		{
			name: "invalid issue ID",
			input: issues.GetIssueByIdInput{
				ID:        "invalid UUID",
				ProjectID: validProjectID.String(),
			},
			wantErr: issues.ErrInvalidIssueID,
		},
		{
			name: "invalid project ID",
			input: issues.GetIssueByIdInput{
				ID:        validIssueID.String(),
				ProjectID: "invalid UUID",
			},
			wantErr: common.ErrInvalidProjectID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
					t.Fatal("GetIssueById should not be called with invalid params")
					return db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.GetIssueByID(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, issues.Issue{}, got)
		})
	}
}

func TestGetIssueById_NotFound(t *testing.T) {
	issueID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{}, pgx.ErrNoRows
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.ErrorIs(t, err, issues.ErrIssueNotFound)
	require.Equal(t, issues.Issue{}, got)
}

func TestGetIssueById_StoreError(t *testing.T) {
	issueID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	expectedErr := errors.New("generic error")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, issues.Issue{}, got)
}
