# metrics

Import path: `github.com/InsideGallery/core/metrics`

`metrics` provides backend-agnostic service instrumentation. Services record counts, gauges, and distributions through a
`Client`; processor packages register concrete exporters by name.

## Main APIs

- `Config` selects processors.
- `GetEnvConfig(prefix ...string)` reads metrics config, defaulting to the `METRICS` prefix.
- `PrometheusOnly(cfg Config)` collapses any enabled config to Prometheus.
- `Processor` is the exporter interface: `Close`, `Count`, `Gauge`, and `Distribution`.
- `Register`, `RegisteredProcessors`, and `Factory` manage processor registration.
- `New(cfg Config, service string)` builds a fanout client.
- `Default`, `SetDefault`, and `InstallDefault` manage the process-wide client.
- `NormalizeTags` returns a sorted copy of tags; `TagSet` joins sorted tags with commas.
- `Counter`, `Gauge`, and `Observer` are resolved metric handles; `HandleProvider` is the optional capability of
  resolving one, implemented by `*Client` and by the Prometheus processor.
- `(*Client).CounterHandle`, `GaugeHandle`, and `DistributionHandle` resolve a handle across every configured
  processor.

## Usage

```go
package example

import (
	"errors"

	_ "github.com/InsideGallery/core/metrics/all"

	"github.com/InsideGallery/core/metrics"
)

func recordMetric() (err error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return err
	}

	client, err := metrics.New(cfg, "api")
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	handle := metrics.InstallDefault(client)
	defer func() {
		err = errors.Join(err, handle.Close())
	}()

	return client.Count("requests_total", 1, []string{"status:ok"})
}
```

## Configuration

`GetEnvConfig` reads:

- `METRICS_PROCESSORS`: comma-separated processor names, default `prometheus`.

Processor names are trimmed, lowercased, de-duplicated, and may be split across comma-separated entries. The values
`none`, `off`, and `disabled` disable metrics. Processor-specific environment variables do not select processors; they
only configure a processor after it has been selected and registered.

## Resolved Handles

`Count`, `Gauge`, and `Distribution` take a `(name, tags)` tuple and every processor has to turn it back into a backend
child metric on each record. A caller on a data path records the same bounded set of tuples forever, so it can resolve
the child once instead:

```go
requests, err := client.CounterHandle("requests_total", []string{"op:get", "status:ok"})
if err != nil {
	return err
}

requests.Add(1) // per operation: no tuple, no lookup
```

Against the Prometheus processor, whose memo already makes a repeat record allocation-free, this is 71.9ns -> 6.8ns per
record (`BenchmarkResolvedHandleVersusMemoizedRecord`, medians of 5, one process, 0 allocations in both arms).

`HandleProvider` is deliberately **not** part of `Processor`. A processor that does not implement it keeps working:
`*Client` adapts its `Count`/`Gauge`/`Distribution` methods into a handle, so `datadog`, `otel`, and `statsd` need no
change, and a client with a mix of processors records into all of them through one handle. A resolution error is
returned rather than absorbed — the caller has no handle and can still record through `Count`.

`Counter`, `Gauge`, and `Observer` are **aliases** to interface literals (`interface{ Add(value int64) }` and peers),
not defined types. That is what lets a consumer pinned to a release without this API declare the same literal locally
and detect the capability with a type assertion, using handles when the linked version provides them and
`Count`/`Distribution` when it does not. Do not turn them into defined types: Go matches method signatures on type
identity, and a defined type is never identical to any other type, so consumers would have to bump their pin in
lockstep.

## Operational Notes

`New` returns `nil, nil` when metrics are disabled. A nil `*Client` is safe to call and returns nil for `Close`,
`Count`, `Gauge`, and `Distribution`.

Import `metrics/all` or the specific processor packages before selecting processor names in `METRICS_PROCESSORS`.
Processor call errors are joined and wrapped with the metric operation and name.
