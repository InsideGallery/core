package logfile

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "LOGFILE"

type Config struct {
	Name  string     `env:"_NAME" envDefault:"/tmp/stderr.log"`
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
