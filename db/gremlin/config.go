package gremlin

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

// EnvPrefix environment prefix for gremlin config
const EnvPrefix = "GREMLIN"

// ConnectionConfig contains required data for gremlin
type ConnectionConfig struct {
	URL string `env:"_URL" envDefault:"ws://127.0.0.1:8182/gremlin"`
}

// GetConnectionConfigFromEnv return aerospike configs bases on environment variables
func GetConnectionConfigFromEnv() (*ConnectionConfig, error) {
	c := new(ConnectionConfig)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: strings.ToUpper(EnvPrefix),
	})
	if err != nil {
		return nil, err
	}

	return c, err
}
