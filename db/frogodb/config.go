package frogodb

import (
	"strings"
	"time"

	fdbclient "github.com/FrogoAI/fdb-client/pkg/client"
	"github.com/caarlos0/env/v10"
)

const (
	// EnvPrefix is the default environment prefix for FrogoDB configuration.
	EnvPrefix = "FDB"

	defaultSeed              = "localhost:3000"
	defaultTendInterval      = 10 * time.Millisecond
	defaultConnectionTimeout = 5 * time.Second
	defaultIdleTimeout       = 55 * time.Second
	defaultPoolSizePerNode   = 64
	defaultMaxConnsPerNode   = 256
	defaultMaxErrorRate      = 100
	defaultErrorRateWindow   = time.Second
)

// ConnectionConfig contains FrogoDB client connection settings.
type ConnectionConfig struct {
	Seeds                    []string      `env:"_SEEDS" envDefault:"localhost:3000" envSeparator:","`
	TendInterval             time.Duration `env:"_TEND_INTERVAL" envDefault:"10ms"`
	ConnectionTimeout        time.Duration `env:"_CONNECTION_TIMEOUT" envDefault:"5s"`
	IdleTimeout              time.Duration `env:"_IDLE_TIMEOUT" envDefault:"55s"`
	PoolSizePerNode          int           `env:"_POOL_SIZE_PER_NODE" envDefault:"64"`
	MaxConnsPerNode          int           `env:"_MAX_CONNS_PER_NODE" envDefault:"256"`
	MaxErrorRate             int           `env:"_MAX_ERROR_RATE" envDefault:"100"`
	ErrorRateWindow          time.Duration `env:"_ERROR_RATE_WINDOW" envDefault:"1s"`
	Multiplexing             bool          `env:"_MULTIPLEXING" envDefault:"false"`
	MultiplexConnsPerNode    int           `env:"_MULTIPLEX_CONNS_PER_NODE" envDefault:"0"`
	MultiplexMinConnsPerNode int           `env:"_MULTIPLEX_MIN_CONNS_PER_NODE" envDefault:"0"`
}

// DefaultConnectionConfig returns FrogoDB connection defaults for the supplied seed addresses.
func DefaultConnectionConfig(seeds ...string) *ConnectionConfig {
	if len(seeds) == 0 {
		seeds = []string{defaultSeed}
	}

	return &ConnectionConfig{
		Seeds:                    seeds,
		TendInterval:             defaultTendInterval,
		ConnectionTimeout:        defaultConnectionTimeout,
		IdleTimeout:              defaultIdleTimeout,
		PoolSizePerNode:          defaultPoolSizePerNode,
		MaxConnsPerNode:          defaultMaxConnsPerNode,
		MaxErrorRate:             defaultMaxErrorRate,
		ErrorRateWindow:          defaultErrorRateWindow,
		Multiplexing:             false,
		MultiplexConnsPerNode:    0,
		MultiplexMinConnsPerNode: 0,
	}
}

// GetConnectionConfigFromEnv returns FrogoDB config from the supplied environment prefix.
func GetConnectionConfigFromEnv(prefix string) (*ConnectionConfig, error) {
	if prefix == "" {
		prefix = EnvPrefix
	}

	config := new(ConnectionConfig)

	if err := env.ParseWithOptions(config, env.Options{Prefix: strings.ToUpper(prefix)}); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *ConnectionConfig) clientConfig() fdbclient.Config {
	config := fdbclient.DefaultConfig(c.Seeds...)
	config.TendInterval = c.TendInterval
	config.ConnectionTimeout = c.ConnectionTimeout
	config.IdleTimeout = c.IdleTimeout
	config.PoolSizePerNode = c.PoolSizePerNode
	config.MaxConnsPerNode = c.MaxConnsPerNode
	config.MaxErrorRate = c.MaxErrorRate
	config.ErrorRateWindow = c.ErrorRateWindow
	config.Multiplexing = c.Multiplexing
	config.MultiplexConnsPerNode = c.MultiplexConnsPerNode
	config.MultiplexMinConnsPerNode = c.MultiplexMinConnsPerNode

	return config
}
