package mongodb

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

var defaultStore = NewClientStore(nil) //nolint:gochecknoglobals // compatibility store

// ClientStore owns a MongoDB client for explicit application composition.
type ClientStore struct {
	mu     sync.RWMutex
	client *MongoClient
}

// NewClientStore creates a MongoDB client store with an optional existing client.
func NewClientStore(client *MongoClient) *ClientStore {
	return &ClientStore{
		client: client,
	}
}

// Set stores a MongoDB client in this store.
func (s *ClientStore) Set(client *MongoClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client = client
}

// Get returns the MongoDB client from this store.
func (s *ClientStore) Get() (*MongoClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return s.client, nil
}

// GetOrCreate returns or creates a MongoDB client from explicit config.
func (s *ClientStore) GetOrCreate(config *ConnectionConfig) (*MongoClient, error) {
	client, err := s.Get()
	if err == nil {
		return client, nil
	}

	if !errors.Is(err, ErrConnectionIsNotSet) {
		return nil, err
	}

	client, err = NewMongoClient(config)
	if err != nil {
		return nil, err
	}

	s.Set(client)

	return client, nil
}

// Close disconnects the stored MongoDB client and clears this store.
func (s *ClientStore) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil || s.client.Client == nil {
		s.client = nil

		return nil
	}

	err := s.client.Disconnect(ctx)
	s.client = nil

	return err
}

// Set global client
//
// Deprecated: use ClientStore.Set on an explicit store.
func Set(r *MongoClient) {
	defaultStore.Set(r)
}

// Get return mongo client
//
// Deprecated: use ClientStore.Get on an explicit store.
func Get() (*MongoClient, error) {
	return defaultStore.Get()
}

// Default return default client
//
// Deprecated: use NewMongoClient or ClientStore.GetOrCreate with explicit config.
func Default() (*MongoClient, error) {
	c, err := Get()
	if err != nil {
		config, err := GetConnectionConfigFromEnv()
		if err != nil {
			slog.Default().Error("Error getting mongo config", "err", err)
			return nil, err
		}

		c, err = defaultStore.GetOrCreate(config)
		if err != nil {
			slog.Default().Error("Error getting mongo client", "err", err)
			return nil, err
		}
	}

	return c, nil
}
