package repository

import (
	"context"
	"errors"
	"time"
)

var ErrTokenNotFound = errors.New("invalid token")

type SessionRepository interface {
	GetToken(c context.Context, key string) (string, error)
	CreateToken(c context.Context, key, value string, ttl time.Duration) error
}
