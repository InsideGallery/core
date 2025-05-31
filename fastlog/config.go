package fastlog

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/go-slog/otelslog"
	slogmulti "github.com/samber/slog-multi"

	"github.com/InsideGallery/core/errors"
	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/fastlog/handlers/nop"
	"github.com/InsideGallery/core/fastlog/middlewares"
)

const EnvPrefix = "LOG"

const (
	Separator = ":"

	typeParts = 2
)

type Config struct {
	Outputs         []string   `env:"_OUTPUTS" envDefault:"stderr:json"`
	Level           slog.Level `env:"_LEVEL" envDefault:"INFO"`
	Caller          bool       `env:"_CALLER" envDefault:"true"`
	ErrorFormatting bool       `env:"_ERROR_FORMATING" envDefault:"false"`
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

func (c *Config) GetHandler(m ...slogmulti.Middleware) (slog.Handler, error) {
	var outputs []slog.Handler
	var errs []error

	for _, out := range c.Outputs {
		parts := strings.Split(out, Separator)
		if len(parts) != typeParts {
			continue
		}
		kind := parts[0]
		format := parts[1]

		o, err := handlers.Get(kind, format, c.Level)
		if err != nil {
			errs = append(errs, err)
		}

		if o == nil {
			continue
		}
		outputs = append(outputs, o)
	}

	if len(outputs) == 0 {
		h, err := handlers.Get(nop.OutKind, handlers.FormatJSON, c.Level)
		if err != nil {
			return nil, errors.Combine(append(errs, err)...)
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
		Handler(otelslog.NewHandler(slogmulti.Fanout(outputs...))), errors.Combine(errs...)
}
