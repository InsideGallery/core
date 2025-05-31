package datadog

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "DATADOG"

type Config struct {
	Host     string        `env:"_HOST" envDefault:""`
	Service  string        `env:"_SERVICE" envDefault:""`
	Endpoint string        `env:"_ENDPOINT" envDefault:"datadoghq.eu"`
	APIKey   string        `env:"_API_KEY" envDefault:""`
	Timeout  time.Duration `env:"_TIMEOUT" envDefault:"5s"`
	Level    slog.Level    `env:"_LEVEL" envDefault:"INFO"`
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
