package fastlog

import (
	"errors"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
)

// SetupDefaultLogger initializes slog.Default() from an explicit configuration.
// Import handler packages (e.g. stderr, otel, datadog) via blank imports to register them.
func SetupDefaultLogger(cfg *Config, m ...slogmulti.Middleware) error {
	if cfg == nil {
		return errors.New("fastlog config is required")
	}

	handler, err := cfg.GetHandler(m...)
	if err != nil {
		return err
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
