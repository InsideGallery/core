package mongodb

import (
	"log/slog"
	"sync"
)

var (
	client *MongoClient
	mu     sync.RWMutex
)

// Set global client
func Set(r *MongoClient) {
	mu.Lock()
	client = r
	mu.Unlock()
}

// Get return mongo client
func Get() (*MongoClient, error) {
	mu.RLock()
	defer mu.RUnlock()

	if client == nil {
		return nil, ErrConnectionIsNotSet
	}

	return client, nil
}

// Default return default client
func Default() (*MongoClient, error) {
	c, err := Get()
	if err != nil {
		config, err := GetConnectionConfigFromEnv()
		if err != nil {
			slog.Default().Error("Error getting mongo config", "err", err)
			return nil, err
		}

		c, err = NewMongoClient(config)
		if err != nil {
			slog.Default().Error("Error getting mongo client", "err", err)
			return nil, err
		}

		Set(c)
	}

	return c, nil
}
