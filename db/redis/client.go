package redis

import (
	"log/slog"
	"sync"
)

var (
	client *Connection
	mu     sync.RWMutex
)

// Set global client
func Set(r *Connection) {
	mu.Lock()
	client = r
	mu.Unlock()
}

// Get return redis client
func Get() (*Connection, error) {
	mu.RLock()
	defer mu.RUnlock()

	if client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return client, nil
}

// Default return default client
func Default() (*Connection, error) {
	c, err := Get()
	if err != nil {
		config, err := GetConnectionConfigFromEnv()
		if err != nil {
			slog.Default().Error("Error getting redis config", "err", err)
			return nil, err
		}

		c = NewRedisClient(config)
		Set(c)
	}

	return c, nil
}
