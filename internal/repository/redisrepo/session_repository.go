package redisrepo

import (
	"context"
	"errors"
	"time"

	"github.com/myacey/avito-shop/internal/repository"
	"github.com/redis/go-redis/v9"
)

type RedisSessionRepository struct {
	rdb *redis.Client
}

func NewRedisSessionRepo(rdb *redis.Client) repository.SessionRepository {
	return &RedisSessionRepository{rdb}
}

func (r *RedisSessionRepository) GetToken(c context.Context, key string) (string, error) {
	tok, err := r.rdb.Get(c, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", repository.ErrTokenNotFound
	}

	return tok, err
}

func (r *RedisSessionRepository) CreateToken(c context.Context, key, value string, ttl time.Duration) error {
	return r.rdb.Set(c, key, value, ttl).Err()
}
