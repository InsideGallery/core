# fastlog/handlers/stderr

Import path: `github.com/InsideGallery/core/fastlog/handlers/stderr`

`stderr` registers a structured log output that writes to `os.Stderr`. It is the default output selected by
`fastlog.Config` when `LOG_OUTPUTS` is unset and the handler has been registered.

## Main APIs

- `OutKind` is the registry key: `stderr`.
- `New()` returns `os.Stderr` and handler options using the configured stderr level.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"

	"github.com/InsideGallery/core/fastlog"
)

func configureLogging() error {
	cfg, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return err
	}

	return fastlog.SetupDefaultLogger(cfg)
}
```

Set `LOG_OUTPUTS=stderr:json` or `LOG_OUTPUTS=stderr:text` to choose the slog encoding.

## Configuration

The package reads the `STDERR` prefix:

- `STDERR_LEVEL`: stderr handler level, default `INFO`.

If `STDERR_LEVEL` cannot be parsed, `New` still returns `os.Stderr` with nil options. When used through the registry,
the registry then applies the level passed by `fastlog.Config`.
