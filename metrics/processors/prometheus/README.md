# metrics/processors/prometheus

Import path: `github.com/InsideGallery/core/metrics/processors/prometheus`

This package registers the Prometheus metrics processor. It records metrics in an in-process Prometheus registry and
exposes the active registry through `HTTPHandler`.

## Main APIs

- `ProcessorName` is the registration name: `prometheus`.
- `New(cfg metrics.Config, service string)` creates the processor and is registered from `init`.
- `HTTPHandler(w http.ResponseWriter, r *http.Request)` serves the active scrape response.

## Usage

```go
package main

import (
	"net/http"

	_ "github.com/InsideGallery/core/metrics/processors/prometheus"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/metrics/processors/prometheus"
)

func newMetrics() (*metrics.Client, error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/metrics", prometheus.HTTPHandler)

	return metrics.New(cfg, "api")
}
```

`METRICS_PROCESSORS` defaults to `prometheus`, but the processor package still has to be imported directly or through
`metrics/all` so it can register itself.

## Configuration

The package reads the `METRICS_PROMETHEUS` prefix for histogram tuning:

- `METRICS_PROMETHEUS_CLASSIC_BUCKETS`: comma-separated finite positive bucket values, sorted and de-duplicated.
- `METRICS_PROMETHEUS_NATIVE_BUCKET_FACTOR`: native histogram bucket factor, default `1.1`; must be greater than `1`.
- `METRICS_PROMETHEUS_NATIVE_ZERO_THRESHOLD`: default `0`.
- `METRICS_PROMETHEUS_NATIVE_MAX_BUCKETS`: default `160`.
- `METRICS_PROMETHEUS_NATIVE_MIN_RESET_DURATION`: default `1h`.
- `METRICS_PROMETHEUS_NATIVE_MAX_ZERO_THRESHOLD`: default `0`.

## Operational Notes

`New` registers Go runtime and process collectors with a constant `service` label, then makes that processor active for
`HTTPHandler`. The latest created processor is active. Closing an inactive processor does not clear the active one;
closing the active processor clears it.

Counts become counters and reject negative values. Gauges become gauges. Distributions become histograms. Tags in
`key:value` form become labels after normalization and sanitization; loose tags are ignored by this processor. When no
processor is active, `HTTPHandler` returns `200 OK` with an empty Prometheus text response.
