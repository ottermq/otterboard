package issues_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
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
