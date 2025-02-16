package redisrepo

import (
	"context"
	"time"

	"github.com/myacey/avito-shop/internal/backconfig"
	"github.com/redis/go-redis/v9"
)

func ConfigureRedisClient(config *backconfig.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost + ":6379",
		Password: "",
		DB:       0,

		PoolSize:        1300,
		MinIdleConns:    550,
		MaxIdleConns:    1000,
		ConnMaxIdleTime: 30 * time.Minute,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
