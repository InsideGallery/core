package aerospike

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

// EnvPrefix environment prefix for mongodb config
const (
	EnvPrefix = "AEROSPIKE"
)

// ConnectionConfig contains required data for mongo
type ConnectionConfig struct {
	Host                string   `env:"_HOST" envDefault:"127.0.0.1"`
	Username            string   `env:"_USERNAME" envDefault:""`
	Password            string   `env:"_PASSWORD" envDefault:""`
	Hosts               []string `env:"_HOSTS" envDefault:""`
	Port                int      `env:"_PORT" envDefault:"3000"`
	ConnectionQueueSize int      `env:"_CONNECTION_QUEUE_SIZE" envDefault:"1000"`
}

// GetConnectionConfigFromEnv return aerospike configs bases on environment variables
func GetConnectionConfigFromEnv(prefix string) (*ConnectionConfig, error) {
	c := new(ConnectionConfig)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: strings.ToUpper(prefix),
	})
	if err != nil {
		return nil, err
	}

	return c, err
}
