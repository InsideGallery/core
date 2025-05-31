package otel

import (
	"log/slog"

	"github.com/agoda-com/opentelemetry-go/otelslog"
	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "OTEL"

type Config struct {
	ServiceName    string     `env:"_SERVICE_NAME" envDefault:"fastlog"`
	ServiceVersion string     `env:"_SERVICE_VERSION" envDefault:"v1.0.0"`
	Namespace      string     `env:"_NAMESPACE" envDefault:"none"`
	Level          slog.Level `env:"_LEVEL" envDefault:"INFO"`
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

func (c *Config) GetOptions() *otelslog.HandlerOptions {
	return &otelslog.HandlerOptions{
		Level: c.Level,
	}
}
