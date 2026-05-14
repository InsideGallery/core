# fastlog

Import path: `github.com/InsideGallery/core/fastlog`

`fastlog` builds and installs structured `log/slog` handlers from environment-backed configuration. It fans out to
registered output handlers, applies built-in middleware, and can replace `slog.Default()`.

## Main APIs

- `Config` describes logging outputs, level, and middleware toggles.
- `GetConfigFromEnv()` reads `LOG_*` environment variables.
- `(*Config).GetHandler(m ...slogmulti.Middleware)` builds a composite `slog.Handler`.
- `SetupDefaultLogger(cfg *Config, m ...slogmulti.Middleware)` installs the handler as `slog.Default()`.

## Usage

```go
package example

import (
	"log/slog"

	_ "github.com/InsideGallery/core/fastlog/all"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/middlewares"
)

func configureLogging() error {
	cfg, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return err
	}

	if err := fastlog.SetupDefaultLogger(cfg, middlewares.NewGDPRMiddleware()); err != nil {
		return err
	}

	slog.Info("logging configured")

	return nil
}
```

## Configuration

`GetConfigFromEnv` uses the `LOG` prefix:

- `LOG_OUTPUTS`: comma-separated `kind:format` values, default `stderr:json`.
- `LOG_LEVEL`: parsed as a `slog.Level`, default `INFO`.
- `LOG_CALLER`: adds a `caller` attribute when true, default `true`.
- `LOG_ERROR_FORMATTING`: converts `error` attributes to structured groups when true, default `false`.

Valid formats are `json` and `text`; unknown formats fall back to JSON. Malformed output entries are skipped. Unknown
handlers are collected as errors, and if no handler can be built the package falls back to the registered `nop` handler.

The base package imports only the `nop` fallback directly. Import `fastlog/all` or the specific handler packages before
selecting `stderr`, `datadog`, or `otel` through `LOG_OUTPUTS`.
