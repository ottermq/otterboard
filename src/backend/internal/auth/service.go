package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"golang.org/x/crypto/bcrypt"
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

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAuthService(store UserStore, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (string, error) {
	_, err := s.store.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return "", ErrEmailAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})

	if err != nil {
		return "", err
	}
	mappedUser := mapDbUserToAuth(user)

	return s.generateAccessToken(mappedUser)

}

func (s *AuthService) generateAccessToken(user User) (string, error) {
	userJson, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"user":  string(userJson),
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
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
