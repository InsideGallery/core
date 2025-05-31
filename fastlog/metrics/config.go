package metrics

import (
	"github.com/caarlos0/env/v10"
	"go.opentelemetry.io/otel/metric"
)

const EnvPrefix = "OTEL"

type Config struct {
	Prefix      string `env:"_PREFIX" envDefault:"fl"`
	Namespace   string `env:"_NAMESPACE" envDefault:"none"`
	ServiceName string `env:"_SERVICE_NAME" envDefault:"core"`
	Version     string `env:"_SERVICE_VERSION" envDefault:"v1.0.0"`
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

func (c *Config) GetOptions() []metric.MeterOption {
	return nil
}
