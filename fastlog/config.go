package fastlog

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v10"
	slogmulti "github.com/samber/slog-multi"

	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/fastlog/handlers/nop"
	"github.com/InsideGallery/core/fastlog/middlewares"
)

const (
	envPrefix = "LOG"
	separator = ":"
	typeParts = 2
)

// Config holds logging configuration parsed from environment variables.
type Config struct {
	Outputs         []string   `env:"_OUTPUTS" envDefault:"stderr:json"`
	Level           slog.Level `env:"_LEVEL" envDefault:"INFO"`
	Caller          bool       `env:"_CALLER" envDefault:"true"`
	ErrorFormatting bool       `env:"_ERROR_FORMATTING" envDefault:"false"`
}

// GetConfigFromEnv reads logging configuration from environment variables.
func GetConfigFromEnv() (*Config, error) {
	c := new(Config)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: envPrefix,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetHandler builds a composite slog.Handler from the configured outputs and middlewares.
func (c *Config) GetHandler(m ...slogmulti.Middleware) (slog.Handler, error) {
	var (
		outputs []slog.Handler
		errs    []error
	)

	for _, out := range c.Outputs {
		parts := strings.Split(out, separator)
		if len(parts) != typeParts {
			continue
		}

		kind := parts[0]
		format := parts[1]

		o, err := handlers.Get(kind, format, c.Level)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if o != nil {
			outputs = append(outputs, o)
		}
	}

	if len(outputs) == 0 {
		h, err := handlers.Get(nop.OutKind, handlers.FormatJSON, c.Level)
		if err != nil {
			return nil, errors.Join(append(errs, err)...)
		}

		outputs = append(outputs, h)
	}

	var orderedMiddlewares []slogmulti.Middleware

	if c.Caller {
		orderedMiddlewares = append(orderedMiddlewares,
			slogmulti.NewHandleInlineMiddleware(middlewares.CallerMiddleware),
		)
	}

	if c.ErrorFormatting {
		orderedMiddlewares = append(orderedMiddlewares,
			slogmulti.NewHandleInlineMiddleware(middlewares.ErrorFormattingMiddleware),
		)
	}

	orderedMiddlewares = append(orderedMiddlewares, m...)

	return slogmulti.Pipe(orderedMiddlewares...).
		Handler(slogmulti.Fanout(outputs...)), errors.Join(errs...)
}
