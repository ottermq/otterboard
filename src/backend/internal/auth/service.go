package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = common.NewAppError(http.StatusConflict, "email already exists")
	ErrInvalidCredentials = common.NewAppError(http.StatusUnauthorized, "invalid credentials")
)

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthService struct {
	store UserStore
}

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAuthService(store UserStore) *AuthService {
	return &AuthService{
		store: store,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (User, error) {
	_, err := s.store.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return User{}, ErrEmailAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})

	if err != nil {
		return User{}, err
	}
	mappedUser := mapDbUserToAuth(user)

	return mappedUser, nil

}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (User, error) {
	user, err := s.store.GetUserByEmail(ctx, input.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password))
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	mappedUser := mapDbUserToAuth(user)

	return mappedUser, nil
}

func mapDbUserToAuth(user db.User) User {
	mappedUser := User{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
	return mappedUser
}
