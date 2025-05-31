package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Connection struct {
	*redis.Client
}

func NewRedisClient(config *ConnectionConfig) *Connection {
	rdb := redis.NewClient(&redis.Options{
		Addr:     strings.Join([]string{config.Host, config.Port}, ":"),
		Username: config.User,
		Password: config.Pass,
		DB:       config.Database,
	})

	return &Connection{
		Client: rdb,
	}
}

func (c Connection) Get(ctx context.Context, key string) (string, bool, error) {
	result, err := c.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}

	return result, true, err
}

func (c Connection) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

func (c Connection) Stop() error {
	return c.Close()
}
