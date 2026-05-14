package stderr

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

const envPrefix = "STDERR"

type config struct {
	Level slog.Level `env:"_LEVEL" envDefault:"INFO"`
}

func getConfigFromEnv() (*config, error) {
	c := new(config)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: envPrefix,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
