package redis

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "REDIS"

type ConnectionConfig struct {
	Host     string `env:"_HOST" envDefault:"localhost"`
	Port     string `env:"_PORT" envDefault:"6379"`
	User     string `env:"_USER" envDefault:""`
	Pass     string `env:"_PASS" envDefault:""`
	Database int    `env:"_DATABASE" envDefault:"0"`
}

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
