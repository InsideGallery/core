package aerospike

import (
	"fmt"
	"sync"

	aero "github.com/aerospike/aerospike-client-go/v7"
	"github.com/aerospike/aerospike-client-go/v7/utils/buffer"
)

var (
	connection = map[string]*aero.Client{}
	mu         sync.RWMutex
)

func init() {
	// Force disable all architectures in aerospike to support int64
	buffer.Arch64Bits = false
	buffer.Arch32Bits = false
}

// Set set global connection
func Set(name string, r *aero.Client) {
	mu.Lock()
	defer mu.Unlock()
	connection[name] = r
}

// Get return aerospike connection
func Get(name string) (*aero.Client, error) {
	mu.RLock()
	defer mu.RUnlock()

	c, exists := connection[name]
	if !exists {
		return nil, ErrConnectionIsNotSet
	}

	return c, nil
}

// Default return default AerospikeInterface
func Default(names ...string) (*aero.Client, error) {
	name := EnvPrefix

	if len(names) > 0 {
		if names[0] != "" {
			name = names[0]
		}
	}

	c, err := Get(name)
	if err != nil {
		config, err := GetConnectionConfigFromEnv(name)
		if err != nil {
			return nil, fmt.Errorf("error getting aerospike config: %w", err)
		}

		if len(config.Hosts) > 0 {
			var hosts []*aero.Host

			hosts, err = aero.NewHosts(config.Hosts...)
			if err != nil {
				return nil, fmt.Errorf("error parse aerospike hosts: %w", err)
			}

			c, err = aero.NewClientWithPolicyAndHost(CreateClientPolicy(config), hosts...)
		} else {
			c, err = aero.NewClientWithPolicy(CreateClientPolicy(config), config.Host, config.Port)
		}

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
