package bunt

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

// EncPrefixDB contains profix of DB
const EncPrefixDB = "DB"

// ConnectionConfig contains listen url for the server and additional options
type ConnectionConfig struct {
	Filename string `env:"_FILENAME" envDefault:":memory:"`
}

// GetConnectionConfigFromEnv return server configs bases on environment variables
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
