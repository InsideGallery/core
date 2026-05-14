# metrics/processors/otel

Import path: `github.com/InsideGallery/core/metrics/processors/otel`

This package registers the OpenTelemetry metrics processor. It records metrics through the global OpenTelemetry meter
provider and does not configure an exporter by itself.

## Main APIs

- `ProcessorName` is the registration name: `otel`.
- `New(cfg metrics.Config, service string)` creates the processor and is registered from `init`.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/metrics/processors/otel"

	"github.com/InsideGallery/core/metrics"
)

func newMetrics() (*metrics.Client, error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, err
	}

	return metrics.New(cfg, "api")
}
```

Run with `METRICS_PROCESSORS=otel` after configuring the process OpenTelemetry meter provider.

## Configuration

The package reads the `METRICS_OTEL` prefix:

- `METRICS_OTEL_METER_NAME`: meter name, default `github.com/InsideGallery/core/metrics`.

Blank meter names fall back to the default. Metric names are sanitized before instrument creation. Attributes always
include `service=<service>` plus sorted tags. Tags in `key:value` form become attributes, spaces in keys become
underscores, and loose tags are recorded under the `tag` attribute key.

Live integration tests are gated by `PTOLEMY_METRICS_OTEL_INTEGRATION=1`.
