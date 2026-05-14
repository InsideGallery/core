# metrics/processors/datadog

Import path: `github.com/InsideGallery/core/metrics/processors/datadog`

This package registers the Datadog DogStatsD metrics processor. It sends count, gauge, and distribution metrics through
`github.com/DataDog/datadog-go/v5/statsd`.

## Main APIs

- `ProcessorName` is the registration name: `datadog`.
- `New(cfg metrics.Config, service string)` creates the processor and is registered from `init`.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/metrics/processors/datadog"

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

Run with `METRICS_PROCESSORS=datadog` and configure an address.

## Configuration

The package reads the `METRICS_DATADOG` prefix:

- `METRICS_DATADOG_ADDR`: DogStatsD address. Required unless `DD_STATSD_ADDR` is set.
- `METRICS_DATADOG_NAMESPACE`: metric namespace, default `ptolemy`.
- `DD_STATSD_ADDR`: legacy fallback address when `METRICS_DATADOG_ADDR` is blank.

The namespace is trimmed and normalized to include one trailing dot. The processor adds a `service:<service>` tag at
client construction and forwards per-metric tags unchanged.

Live integration tests require `PTOLEMY_METRICS_DATADOG_INTEGRATION=1` and one address variable.
