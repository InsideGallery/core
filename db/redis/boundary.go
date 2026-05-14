// Package redis provides Redis connection and key-value helpers.
//
// New code should use explicit connection ownership and core-owned command
// shapes:
//
//	import "github.com/InsideGallery/core/db/redis"
//
//	store := redis.NewConnectionStore(redis.NewRedisClient(config))
//	conn, err := store.Get()
//
// Prefer KeyValueStore with GetOptions, SetOptions, StringResult, and
// CommandResult for application-facing code.
//
// Compatibility: package-level Set, Get, and Default remain available for
// existing consumers. Prefer NewRedisClient or ConnectionStore.GetOrCreate with
// explicit configuration in new code.
package redis

import (
	"context"
	"time"

	coreerrors "github.com/InsideGallery/core/errors"
)

// GetOptions is the core-owned input for reading a Redis strings value.
type GetOptions struct {
	Key string
}

// SetOptions is the core-owned input for writing a Redis strings value.
type SetOptions struct {
	Key   string
	Value string
	TTL   time.Duration
}

// StringResult is the core-owned result for Redis strings reads.
type StringResult struct {
	Key     string
	Value   string
	Present bool
}

// CommandResult is the core-owned result for Redis commands.
type CommandResult struct {
	Key string
}

// KeyValueStore is the core-owned Redis contract for new consumers.
type KeyValueStore interface {
	GetValue(ctx context.Context, options GetOptions) (StringResult, error)
	SetValue(ctx context.Context, options SetOptions) (CommandResult, error)
	Stop() error
}

// GetValue reads a Redis strings value with core-owned options.
func (c Connection) GetValue(ctx context.Context, options GetOptions) (StringResult, error) {
	value, present, err := c.Get(ctx, options.Key)
	if err != nil {
		return StringResult{}, coreerrors.WrapBoundary("redis", "get value", err)
	}

	return StringResult{
		Key:     options.Key,
		Value:   value,
		Present: present,
	}, nil
}

// SetValue writes a Redis strings value with core-owned options.
func (c Connection) SetValue(ctx context.Context, options SetOptions) (CommandResult, error) {
	if err := c.Set(ctx, options.Key, options.Value, options.TTL); err != nil {
		return CommandResult{}, coreerrors.WrapBoundary("redis", "set value", err)
	}

	return CommandResult{Key: options.Key}, nil
}
