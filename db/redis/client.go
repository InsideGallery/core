package redis

import (
	"errors"
	"log/slog"
	"sync"
)

var defaultStore = NewConnectionStore(nil) //nolint:gochecknoglobals // compatibility store

// ConnectionStore owns a Redis client for explicit application composition.
type ConnectionStore struct {
	mu     sync.RWMutex
	client *Connection
}

// NewConnectionStore creates a Redis connection store with an optional existing client.
func NewConnectionStore(client *Connection) *ConnectionStore {
	return &ConnectionStore{
		client: client,
	}
}

// Set stores a Redis connection in this store.
func (s *ConnectionStore) Set(client *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client = client
}

// Get returns the Redis connection from this store.
func (s *ConnectionStore) Get() (*Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return s.client, nil
}

// GetOrCreate returns or creates a Redis connection from explicit config.
func (s *ConnectionStore) GetOrCreate(config *ConnectionConfig) (*Connection, error) {
	client, err := s.Get()
	if err == nil {
		return client, nil
	}

	if !errors.Is(err, ErrConnectionIsNotSet) {
		return nil, err
	}

	client = NewRedisClient(config)
	s.Set(client)

	return client, nil
}

// Close closes the stored Redis connection and clears this store.
func (s *ConnectionStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil
	}

	err := s.client.Stop()
	s.client = nil

	return err
}

// Set global client
//
// Deprecated: use ConnectionStore.Set on an explicit store.
func Set(r *Connection) {
	defaultStore.Set(r)
}

// Get return redis client
//
// Deprecated: use ConnectionStore.Get on an explicit store.
func Get() (*Connection, error) {
	return defaultStore.Get()
}

// Default return default client
//
// Deprecated: use NewRedisClient or ConnectionStore.GetOrCreate with explicit config.
func Default() (*Connection, error) {
	c, err := Get()
	if err != nil {
		config, err := GetConnectionConfigFromEnv()
		if err != nil {
			slog.Default().Error("Error getting redis config", "err", err)
			return nil, err
		}

		c, err = defaultStore.GetOrCreate(config)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}
