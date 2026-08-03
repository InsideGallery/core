# metrics/processors/prometheus

Import path: `github.com/InsideGallery/core/metrics/processors/prometheus`

This package registers the Prometheus metrics processor. It records metrics in an in-process Prometheus registry and
exposes the active registry through `HTTPHandler`.

## Main APIs

- `ProcessorName` is the registration name: `prometheus`.
- `New(cfg metrics.Config, service string)` creates the processor and is registered from `init`.
- `HTTPHandler(w http.ResponseWriter, r *http.Request)` serves the active scrape response.
- The processor implements `metrics.HandleProvider`: `CounterHandle`, `GaugeHandle`, and `DistributionHandle`
  return the resolved child metric for a tuple.

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

## Resolved Handles

The memo below removes label resolution from a repeat record, but not the lookup that finds the memoized child: the key
is hashed on every record. A caller that records one tuple for the life of the process can resolve the child once
through `metrics.HandleProvider` and keep it — 71.9ns -> 6.8ns per record on the same processor and tuple
(`BenchmarkResolvedHandleVersusMemoizedRecord`, medians of 5, one process, 0 allocations in both arms).

Resolving a handle deliberately does **not** populate the memo: a handle never looks its tuple up again, so an entry
for it would occupy one of the bounded slots without ever being read, and once the memo is full it would ration
admission against tuples that do read it. A handle records into exactly the series `Count`/`Gauge`/`Distribution` would
have produced for the same tuple. Because a handle has no error channel, a negative counter increment is dropped rather
than reported — `Count` returns an error for it, and `stdprom.Counter.Add` panics on it.

## Recording Cost

The processor is safe to call from a data path. The first record of a `(name, tags)` tuple resolves its labels and
registers the collector; every later record of that tuple reuses the resolved child metric through a lock-free memo, so
it costs one map lookup plus the increment and allocates nothing (~65ns against ~1µs and 10 allocations without the
memo). Two cases fall back to resolving labels per record, which is correct but not allocation-free:

- tag lists wider than six entries, the fixed-size memo key;
- tuples the memo turns away once it holds 4096 entries, the bound that keeps an unintended high-cardinality label
  (a request ID, a timestamp) from growing the memo without limit.

Emitted metric names, label names, and label values are identical either way. Prefer stable, bounded label values for
anything recorded per request.

Reaching the bound is not permanent. A full memo admits one turned-away tuple every 512 attempts, evicting an
arbitrary resident to make room, so a cardinality burst costs the tuples it displaces per-call label resolution until
they are recorded often enough to be re-admitted — not for the life of the process. Admission is rationed because
publication is copy-on-write: every admission rebuilds the whole map, and admitting every first-seen tuple would turn
an unbounded label into a 176µs recording path against 465ns for the rationed policy and 153ns for refusing outright
(`BenchmarkHandleCachePolicy/oversized_cycle`). Since attempts are what wins admission, the tuples recorded most often
come back first. Recording is correct throughout: an evicted tuple keeps its Prometheus child and its accumulated
value, and re-resolves to the same series.

## Self-Instrumentation

Both fallbacks are silent — recording stays correct, it just costs what it cost before the memo existed — so the
processor exports its own cache state on every scrape. No caller wiring is required; the collector is registered in
`New` and reads the caches at scrape time, so being observable costs the recording path nothing.

- `metrics_handle_cache_size{kind="counter|gauge|histogram"}` (gauge): handles currently memoized for that cache.
- `metrics_handle_cache_bypass_total{kind,reason="width|full"}` (counter): records that bypassed the memo, split by
  cause — `width` for tag lists wider than six entries, `full` for records of a tuple a full cache turned away.
- `metrics_handle_cache_evictions_total{kind}` (counter): memoized handles dropped to admit a turned-away tuple.

All three carry the same constant `service` label as every other series from this processor.

`metrics_handle_cache_size` reaching 4096 is the condition worth alerting on: the memo is full, and
`bypass_total{reason="full"}` and `evictions_total` are both climbing — the cache is churning, and the tuples in flight
resolve their labels on every record until they are re-admitted. That points at an unintended high-cardinality label
value; the fix is at the emission site, not here. `evictions_total` climbing while `bypass_total{reason="full"}` is flat
means a working set slightly larger than the bound rather than an unbounded label. A non-zero `reason="width"` is
milder and static — some call site records more than six tags and always will.
