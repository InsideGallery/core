//go:generate mockgen -source=interface.go -destination=mocks/client.go
package redis

import (
	"context"
	"time"
)

type Client interface {
	Get(ctx context.Context, key string) (result string, present bool, err error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Stop() error
}
