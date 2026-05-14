# fastlog/middlewares

Import path: `github.com/InsideGallery/core/fastlog/middlewares`

`middlewares` contains reusable `slog` middleware for fastlog pipelines.

## Main APIs

- `CallerMiddleware` adds a `caller` attribute in `file:line` form while preserving existing attributes.
- `ErrorFormattingMiddleware` converts an `error` attribute holding an `error` value into a group with `type` and
  `message` fields.
- `NewGDPRMiddleware()` returns a handler middleware that masks exact keys `password`, `email`, and `phone` with
  `*******`.

## Usage

```go
package example

import (
	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/middlewares"
)

func configureLogging() error {
	cfg, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return err
	}

	return fastlog.SetupDefaultLogger(cfg, middlewares.NewGDPRMiddleware())
}
```

## Operational Notes

`fastlog.Config` wires `CallerMiddleware` from `LOG_CALLER` and `ErrorFormattingMiddleware` from
`LOG_ERROR_FORMATTING`. Custom middleware passed to `GetHandler` or `SetupDefaultLogger` runs after those built-in
middlewares.

The GDPR middleware masks matching top-level attributes, attributes added through `WithAttrs`, and nested group
attributes when the group name or attribute key is one of the configured PII keys.
