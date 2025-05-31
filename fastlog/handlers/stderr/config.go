package stderr

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "STDERR"

type Config struct {
	Level slog.Level `env:"_LEVEL" envDefault:"INFO"`
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
