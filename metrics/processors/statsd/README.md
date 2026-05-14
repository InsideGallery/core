# metrics/processors/statsd

Import path: `github.com/InsideGallery/core/metrics/processors/statsd`

This package registers a plain UDP StatsD metrics processor. It writes count, gauge, and timing packets to a configured
StatsD address.

## Main APIs

- `ProcessorName` is the registration name: `statsd`.
- `New(cfg metrics.Config, service string)` creates the processor and is registered from `init`.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/metrics/processors/statsd"

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

Run with `METRICS_PROCESSORS=statsd` and configure `METRICS_STATSD_ADDR`.

## Configuration

The package reads the `METRICS_STATSD` prefix:

- `METRICS_STATSD_ADDR`: UDP StatsD address. Required.
- `METRICS_STATSD_NAMESPACE`: metric namespace, default `ptolemy`.

The namespace is trimmed and normalized to include one trailing dot. Metric names are sanitized before writing. Tags are
not encoded in StatsD packets by this processor. `Distribution` values are sent as `ms` timing packets.

Live integration tests require `PTOLEMY_METRICS_STATSD_INTEGRATION=1` and `METRICS_STATSD_ADDR`.
