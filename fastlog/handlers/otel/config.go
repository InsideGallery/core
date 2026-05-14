package otel

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

const envPrefix = "OTEL"

type config struct {
	ServiceName    string     `env:"_SERVICE_NAME"    envDefault:"ptolemy"`
	ServiceVersion string     `env:"_SERVICE_VERSION" envDefault:"v1.0.0"`
	Namespace      string     `env:"_NAMESPACE"       envDefault:"default"`
	Level          slog.Level `env:"_LEVEL"           envDefault:"INFO"`
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
