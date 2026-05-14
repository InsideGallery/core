# fastlog/handlers/datadog

Import path: `github.com/InsideGallery/core/fastlog/handlers/datadog`

This package registers the `datadog` fastlog output. It builds a Datadog `slog` handler using the Datadog API client and
the `slog-datadog` adapter.

## Main APIs

- `OutKind` is the registry key: `datadog`.

The handler factory is registered from `init`, so most consumers use this package as a blank import or through
`fastlog/all`.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/datadog"

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

Run with `LOG_OUTPUTS=datadog:json`. The registry format field is ignored because this package registers a complete
handler factory.

## Configuration

The package reads the `DATADOG` prefix:

- `DATADOG_HOST`: parsed by config; the current handler uses `os.Hostname()` for the emitted hostname.
- `DATADOG_SERVICE`: service name sent to Datadog.
- `DATADOG_ENDPOINT`: Datadog site, default `datadoghq.eu`.
- `DATADOG_API_KEY`: API key.
- `DATADOG_TIMEOUT`: API timeout, default `5s`.
- `DATADOG_LEVEL`: handler level, default `INFO`.

Live integration tests require `PTOLEMY_FASTLOG_DATADOG_INTEGRATION=1` and `DATADOG_API_KEY`.
