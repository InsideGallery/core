package metrics //nolint:revive // package name intentionally matches directory/domain usage

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

const envPrefix = "METRICS"

// Config holds backend-agnostic metrics configuration.
type Config struct {
	Processors []string `env:"_PROCESSORS" envDefault:"prometheus"`
}

// Enabled reports whether any processor is configured.
func (c Config) Enabled() bool {
	return len(c.EnabledProcessors()) > 0
}

// EnabledProcessors returns the configured processors.
func (c Config) EnabledProcessors() []string {
	return normalizeProcessors(c.Processors)
}

// GetEnvConfig reads metrics configuration from environment variables.
// Default prefix is METRICS. Processor-specific packages own their own env config.
func GetEnvConfig(prefix ...string) (Config, error) {
	p := envPrefix
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}

	var cfg Config

	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: strings.ToUpper(p),
	}); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func normalizeProcessors(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(raw))
	processors := make([]string, 0, len(raw))

	for _, entry := range raw {
		for _, part := range strings.Split(entry, ",") {
			processor := strings.ToLower(strings.TrimSpace(part))
			if processor == "" || isDisabledProcessor(processor) {
				continue
			}

			if _, ok := seen[processor]; ok {
				continue
			}

			seen[processor] = struct{}{}
			processors = append(processors, processor)
		}
	}

	return processors
}

func isDisabledProcessor(processor string) bool {
	return processor == "none" || processor == "off" || processor == "disabled"
}
