package logstash

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "LOGSTASH"

type Config struct {
	Host    string     `env:"_HOST" envDefault:"localhost:4242"`
	Network string     `env:"_NETWORK" envDefault:"tcp"`
	Level   slog.Level `env:"_LEVEL" envDefault:"INFO"`
}

func GetConfigFromEnv() (*Config, error) {
	c := new(Config)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: EnvPrefix,
	})
	if err != nil {
		return nil, err
	}

	return c, err
}
