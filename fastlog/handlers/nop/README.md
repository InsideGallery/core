# fastlog/handlers/nop

Import path: `github.com/InsideGallery/core/fastlog/handlers/nop`

`nop` registers a no-op fastlog output. It is useful for tests, disabled logging, and as the fallback when `fastlog`
cannot build any configured output.

## Main APIs

- `OutKind` is the registry key: `nop`.
- `W` is an `io.Writer` that discards bytes and reports the full byte count as written.
- `New()` returns a `W` writer and empty handler options.

## Usage

```go
package example

import (
	"log/slog"

	"github.com/InsideGallery/core/fastlog"
)

func configureQuietLogger() error {
	cfg := &fastlog.Config{
		Outputs: []string{"nop:json"},
		Level:   slog.LevelInfo,
	}

	return fastlog.SetupDefaultLogger(cfg)
}
```

## Operational Notes

There is no environment specific to this package. The root `fastlog` package imports `nop` directly so the fallback is
available even when `fastlog/all` is not imported.
