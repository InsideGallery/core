package aerospike

import (
	"errors"
	"fmt"
	"sync"

	aero "github.com/aerospike/aerospike-client-go/v7"
	"github.com/aerospike/aerospike-client-go/v7/utils/buffer"
)

var defaultConnections = NewConnectionRegistry() //nolint:gochecknoglobals // compatibility registry

// BufferArchitecture exposes Aerospike buffer architecture flags as explicit dependencies.
type BufferArchitecture struct {
	Arch64Bits *bool
	Arch32Bits *bool
}

func init() {
	// Deprecated: call Setup with explicit BufferArchitecture dependencies.
	Setup(DefaultBufferArchitecture())
}

// DefaultBufferArchitecture returns the Aerospike package buffer architecture flags.
func DefaultBufferArchitecture() BufferArchitecture {
	return BufferArchitecture{
		Arch64Bits: &buffer.Arch64Bits,
		Arch32Bits: &buffer.Arch32Bits,
	}
}

// Setup disables Aerospike buffer architecture flags for int64 compatibility.
func Setup(architecture BufferArchitecture) {
	disableBufferArchitectureFlag(architecture.Arch64Bits)
	disableBufferArchitectureFlag(architecture.Arch32Bits)
}

func disableBufferArchitectureFlag(flag *bool) {
	if flag == nil {
		return
	}

	*flag = false
}

// ConnectionRegistry owns Aerospike clients for explicit application composition.
type ConnectionRegistry struct {
	mu          sync.RWMutex
	connections map[string]*aero.Client
}

// NewConnectionRegistry creates an isolated Aerospike connection registry.
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]*aero.Client),
	}
}

// Set stores an Aerospike client in this registry.
func (r *ConnectionRegistry) Set(name string, client *aero.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connections[name] = client
}

// Get returns an Aerospike client from this registry.
func (r *ConnectionRegistry) Get(name string) (*aero.Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.connections[name]
	if !exists {
		return nil, ErrConnectionIsNotSet
	}

	return client, nil
}

// Default returns or creates a registry-scoped Aerospike client from explicit config.
func (r *ConnectionRegistry) Default(name string, config *ConnectionConfig) (*aero.Client, error) {
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
			client.Close()
		}

		delete(r.connections, name)
	}

	return nil
}

// NewConnection creates an Aerospike client from explicit config.
func NewConnection(config *ConnectionConfig) (*aero.Client, error) {
	var (
		client *aero.Client
		err    error
	)

	if len(config.Hosts) > 0 {
		var hosts []*aero.Host

		hosts, err = aero.NewHosts(config.Hosts...)
		if err != nil {
			return nil, fmt.Errorf("parse aerospike hosts: %w", err)
		}

		client, err = aero.NewClientWithPolicyAndHost(CreateClientPolicy(config), hosts...)
	} else {
		client, err = aero.NewClientWithPolicy(CreateClientPolicy(config), config.Host, config.Port)
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewConnectionFromEnv creates an Aerospike client from an explicit environment prefix.
func NewConnectionFromEnv(prefix string) (*aero.Client, error) {
	config, err := GetConnectionConfigFromEnv(prefix)
	if err != nil {
		return nil, fmt.Errorf("get aerospike config: %w", err)
	}

	return NewConnection(config)
}

// Set set global connection
//
// Deprecated: use ConnectionRegistry.Set on an explicit registry.
func Set(name string, r *aero.Client) {
	defaultConnections.Set(name, r)
}

// Get return aerospike connection
//
// Deprecated: use ConnectionRegistry.Get on an explicit registry.
func Get(name string) (*aero.Client, error) {
	return defaultConnections.Get(name)
}

// Default return default AerospikeInterface
//
// Deprecated: use NewConnection or ConnectionRegistry.Default with explicit config.
func Default(names ...string) (*aero.Client, error) {
	name := EnvPrefix

	if len(names) > 0 {
		if names[0] != "" {
			name = names[0]
		}
	}

	c, err := Get(name)
	if err != nil {
		c, err = NewConnectionFromEnv(name)
		if err != nil {
			return nil, err
		}

		Set(name, c)
	}

	return c, nil
}

// CreateClientPolicy creates policy for production
func CreateClientPolicy(cfg *ConnectionConfig) *aero.ClientPolicy {
	policy := aero.NewClientPolicy()
	if cfg.Username != "" && cfg.Password != "" {
		policy.User = cfg.Username
		policy.Password = cfg.Password
		policy.AuthMode = aero.AuthModeInternal
	}

	if cfg.ConnectionQueueSize > 0 {
		policy.ConnectionQueueSize = cfg.ConnectionQueueSize
	}

	return policy
}
