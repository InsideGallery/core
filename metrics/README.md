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
- `SeriesDeleter` is the optional capability of retiring a published series, implemented by `*Client` and by the
  Prometheus processor; `(*Client).DeleteSeriesMatching` fans the delete out to whichever processors have it.
- `ErrSeriesDeleteUnsupported` and `ErrEmptySeriesMatch` are the two refusals a delete can report.

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

## Series Retirement

A backend that holds its series in-process keeps exporting one after the last record: the value freezes and the series
stays on every scrape until the process exits. Where the label identifies something that comes and goes — a peer
address, a pod IP, a tenant — that is unbounded cardinality growth in dead series, and a frozen counter reads to an
operator like an active fault. `DeleteSeriesMatching` removes the children themselves:

```go
if err := client.DeleteSeriesMatching("fabric_connection_errors_total", []string{"peer:" + address}); err != nil {
	slog.Warn("retire peer series", "peer", address, "error", err)
}
```

The match is **partial**: `tags` names the labels that identify the subject, and every child carrying them is deleted
whatever its other labels hold. That is deliberate — a series split by an open-ended label (an error `reason`, a status
class) has children the caller cannot enumerate, and retiring the subject has to take all of them. An empty match would
select every child of the metric, so it is refused with `ErrEmptySeriesMatch` rather than obeyed; the case that matters
is not a caller typing `nil` but a caller assembling a tag from an empty subject.

`SeriesDeleter` is **not** part of `Processor`, for the same reason `HandleProvider` is not — but unlike a handle it
cannot be adapted, because a push backend has no resident series to retire and the OpenTelemetry SDK exposes no removal
at all. A processor without the capability is therefore skipped rather than failed. A client where *no* processor can
delete reports `ErrSeriesDeleteUnsupported`, so a caller learns the label it wanted gone is still being exported
instead of assuming success.

Two rules for callers:

- **Retire only what has genuinely gone away.** A counter for a subject that is merely unreachable is exactly the
  signal an operator needs during an incident, and deleting it destroys that signal at the moment it matters. Where
  the same subject can come back at the same identity, treat its series as a counter reset — do not write `absent()`
  alerts on them.
- **Retire series recorded through `Count`, `Gauge`, and `Distribution`.** Those resolve their backend child per record
  and recreate it, so a returning subject counts from zero with no residue. A handle resolved through `HandleProvider`
  holds the child directly, so deleting that child orphans the handle: records through it are accepted and exported
  nowhere. Re-resolve the handle after retiring, or keep a retirable series on the recording path.

## Operational Notes

`New` returns `nil, nil` when metrics are disabled. A nil `*Client` is safe to call and returns nil for `Close`,
`Count`, `Gauge`, and `Distribution`.

Import `metrics/all` or the specific processor packages before selecting processor names in `METRICS_PROCESSORS`.
Processor call errors are joined and wrapped with the metric operation and name.
