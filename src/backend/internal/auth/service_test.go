package auth_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ottermq/otterboard/src/backend/internal/auth"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/stretchr/testify/require"
)

type mockUserStore struct {
	getUserByEmailFn func(ctx context.Context, email string) (db.User, error)
	createUserFn     func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

func (m *mockUserStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return m.getUserByEmailFn(ctx, email)
}

func (m *mockUserStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return m.createUserFn(ctx, arg)
}

func TestRegister_Success(t *testing.T) {
	store := &mockUserStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, pgx.ErrNoRows // email not taken
		},
		createUserFn: func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
			return db.User{
				Email: arg.Email,
				Name:  arg.Name,
			}, nil
		},
	}

	service := auth.NewAuthService(store)
	user, err := service.Register(context.Background(), auth.RegisterInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.Equal(t, "john@example.com", user.Email)
	require.Equal(t, "John Doe", user.Name)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	store := &mockUserStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, nil // user found, no error
		},
	}

	service := auth.NewAuthService(store)
	_, err := service.Register(context.Background(), auth.RegisterInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	})

	require.ErrorIs(t, err, auth.ErrEmailAlreadyExists)
}
