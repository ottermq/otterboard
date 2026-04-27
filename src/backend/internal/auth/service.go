package auth

import (
	"context"
	"errors"

	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
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

type AuthService struct {
	store     UserStore
	jwtSecret string
}

func NewAuthService(store UserStore, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (string, error) {
	panic("not implemented")
}
