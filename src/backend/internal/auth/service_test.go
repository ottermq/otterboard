package auth_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/auth"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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

func TestLogin_Success(t *testing.T) {
	store := &mockUserStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return db.User{
				Email:        "john@example.com",
				PasswordHash: pgtype.Text{String: string(hash), Valid: true},
			}, nil
		},
	}

	service := auth.NewAuthService(store)
	user, err := service.Login(context.Background(), auth.LoginInput{
		Email:    "john@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	require.Equal(t, "john@example.com", user.Email)
}

func TestLogin_UserNotFound(t *testing.T) {
	store := &mockUserStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{}, pgx.ErrNoRows // user not found
		},
	}

	service := auth.NewAuthService(store)
	_, err := service.Login(context.Background(), auth.LoginInput{
		Email:    "john@example.com",
		Password: "whatever",
	})
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	store := &mockUserStore{
		getUserByEmailFn: func(_ context.Context, _ string) (db.User, error) {
			return db.User{
				Email:        "john@example.com",
				PasswordHash: pgtype.Text{String: "$2a$10$examplehash", Valid: true},
			}, nil
		},
	}

	service := auth.NewAuthService(store)
	_, err := service.Login(context.Background(), auth.LoginInput{
		Email:    "john@example.com",
		Password: "wrongpassword",
	})

	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}
