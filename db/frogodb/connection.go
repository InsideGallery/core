package frogodb

import (
	"errors"
	"fmt"
	"sync"

	fdbclient "github.com/FrogoAI/fdb-client/pkg/client"
)

var defaultConnections = NewConnectionRegistry() //nolint:gochecknoglobals // compatibility registry

// ConnectionRegistry owns FrogoDB clients for explicit application composition.
type ConnectionRegistry struct {
	mu          sync.RWMutex
	connections map[string]*fdbclient.Client
}

// NewConnectionRegistry creates an isolated FrogoDB connection registry.
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*fdbclient.Client),
	}
}

// Set stores a FrogoDB client in this registry.
func (r *ConnectionRegistry) Set(name string, client *fdbclient.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connections[name] = client
}

// Get returns a FrogoDB client from this registry.
func (r *ConnectionRegistry) Get(name string) (*fdbclient.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.connections[name]
	if !exists {
		return nil, ErrConnectionIsNotSet
	}

	return client, nil
}

// Default returns or creates a registry-scoped FrogoDB client from explicit config.
func (r *ConnectionRegistry) Default(name string, config *ConnectionConfig) (*fdbclient.Client, error) {
	client, err := r.Get(name)
	if err == nil {
		return client, nil
	}

	if !errors.Is(err, ErrConnectionIsNotSet) {
		return nil, err
	}

	client, err = NewConnection(config)
	if err != nil {
		return nil, err
	}

	r.Set(name, client)

	return client, nil
}

// Close closes all clients stored in this registry and clears it.
func (r *ConnectionRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, client := range r.connections {
		if client != nil {
			if err := client.Close(); err != nil {
				return fmt.Errorf("close frogodb %q: %w", name, err)
			}
		}

		delete(r.connections, name)
	}

	return nil
}

// NewConnection creates a FrogoDB client from explicit config.
func NewConnection(config *ConnectionConfig) (*fdbclient.Client, error) {
	if config == nil {
		return nil, ErrConnectionConfigIsNotSet
	}

	client, err := fdbclient.NewWithConfig(config.clientConfig())
	if err != nil {
		return nil, fmt.Errorf("connect frogodb: %w", err)
	}

	return client, nil
}

// NewConnectionFromEnv creates a FrogoDB client from an explicit environment prefix.
func NewConnectionFromEnv(prefix string) (*fdbclient.Client, error) {
	config, err := GetConnectionConfigFromEnv(prefix)
	if err != nil {
		return nil, fmt.Errorf("get frogodb config: %w", err)
	}

	return NewConnection(config)
}

// Set stores the package-level FrogoDB client.
//
// Deprecated: use ConnectionRegistry.Set on an explicit registry.
func Set(name string, client *fdbclient.Client) {
	defaultConnections.Set(name, client)
}

// Get returns the package-level FrogoDB client.
//
// Deprecated: use ConnectionRegistry.Get on an explicit registry.
func Get(name string) (*fdbclient.Client, error) {
	return defaultConnections.Get(name)
}

// Default returns the package-level FrogoDB client.
//
// Deprecated: use NewConnection or ConnectionRegistry.Default with explicit config.
func Default(names ...string) (*fdbclient.Client, error) {
	name := EnvPrefix

	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}

	client, err := Get(name)
	if err == nil {
		return client, nil
	}

	if !errors.Is(err, ErrConnectionIsNotSet) {
		return nil, err
	}

	client, err = NewConnectionFromEnv(name)
	if err != nil {
		return nil, err
	}

	Set(name, client)

	return client, nil
}
