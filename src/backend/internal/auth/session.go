package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 24 * time.Hour

type SessionStore interface {
	Create(ctx context.Context, userID string) (sessionID string, err error)
	Get(ctx context.Context, sessionID string) (userID string, err error)
	Delete(ctx context.Context, sessionID string) error
}

type redisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(client *redis.Client) SessionStore {
	return &redisSessionStore{client: client}
}

func (s *redisSessionStore) Create(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.NewString()
	key := sessionKey(sessionID)
	if err := s.client.Set(ctx, key, userID, sessionTTL).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

func (s *redisSessionStore) Get(ctx context.Context, sessionID string) (string, error) {
	userID, err := s.client.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (s *redisSessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, sessionKey(sessionID)).Err()
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
