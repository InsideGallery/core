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

## Recording Cost

The processor is safe to call from a data path. The first record of a `(name, tags)` tuple resolves its labels and
registers the collector; every later record of that tuple reuses the resolved child metric through a lock-free memo, so
it costs one map lookup plus the increment and allocates nothing (~65ns against ~1µs and 10 allocations without the
memo). Two cases fall back to resolving labels per record, which is correct but not allocation-free:

- tag lists wider than six entries, the fixed-size memo key;
- tuples first seen after the memo reached 4096 entries, the bound that keeps an unintended high-cardinality label
  (a request ID, a timestamp) from growing the memo without limit.

Emitted metric names, label names, and label values are identical either way. Prefer stable, bounded label values for
anything recorded per request.

## Self-Instrumentation

Both fallbacks are silent — recording stays correct, it just costs what it cost before the memo existed — so the
processor exports its own cache state on every scrape. No caller wiring is required; the collector is registered in
`New` and reads the caches at scrape time, so being observable costs the recording path nothing.

- `metrics_handle_cache_size{kind="counter|gauge|histogram"}` (gauge): handles currently memoized for that cache.
- `metrics_handle_cache_bypass_total{kind,reason="width|full"}` (counter): records that bypassed the memo, split by
  cause — `width` for tag lists wider than six entries, `full` for tuples first seen after the cache reached 4096
  entries.

Both carry the same constant `service` label as every other series from this processor.

`metrics_handle_cache_size` reaching 4096 is the condition worth alerting on: the memo is full, `bypass_total{reason="full"}`
is climbing, and every tuple first seen from then on resolves its labels on every record for the life of the process.
That points at an unintended high-cardinality label value. A non-zero `reason="width"` is milder and static — some
call site records more than six tags and always will.
