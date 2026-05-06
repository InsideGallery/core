package statsd

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

const (
	envPrefix        = "METRICS_STATSD"
	defaultNamespace = "ptolemy"
)

type config struct {
	Addr      string `env:"_ADDR" envDefault:""`
	Namespace string `env:"_NAMESPACE" envDefault:"ptolemy"`
}

func getConfigFromEnv() (config, error) {
	var cfg config

	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: envPrefix,
	}); err != nil {
		return config{}, err
	}

	cfg.Addr = strings.TrimSpace(cfg.Addr)

	cfg.Namespace = strings.TrimSpace(cfg.Namespace)
	if cfg.Namespace == "" {
		cfg.Namespace = defaultNamespace
	}

	return cfg, nil
}

func (c config) namespacePrefix() string {
	return strings.TrimSuffix(c.Namespace, ".") + "."
}
