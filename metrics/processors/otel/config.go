package otel

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

const (
	envPrefix        = "METRICS_OTEL"
	defaultMeterName = "github.com/InsideGallery/core/metrics"
)

type config struct {
	MeterName string `env:"_METER_NAME" envDefault:"github.com/InsideGallery/core/metrics"`
}

func getConfigFromEnv() (config, error) {
	var cfg config

	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: envPrefix,
	}); err != nil {
		return config{}, err
	}

	cfg.MeterName = strings.TrimSpace(cfg.MeterName)
	if cfg.MeterName == "" {
		cfg.MeterName = defaultMeterName
	}

	return cfg, nil
}
