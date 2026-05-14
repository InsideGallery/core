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

## Operational Notes

`New` returns `nil, nil` when metrics are disabled. A nil `*Client` is safe to call and returns nil for `Close`,
`Count`, `Gauge`, and `Distribution`.

Import `metrics/all` or the specific processor packages before selecting processor names in `METRICS_PROCESSORS`.
Processor call errors are joined and wrapped with the metric operation and name.
