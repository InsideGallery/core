//go:build !metrics_minimal

package all_test

import (
	"testing"

	_ "github.com/InsideGallery/core/metrics/all"

	"github.com/InsideGallery/core/metrics"
)

func TestMetricsAllRegistersProcessors(t *testing.T) {
	t.Parallel()

	registered := registeredProcessors(metrics.DefaultRegistry().RegisteredProcessors())
	cases := []struct {
		name      string
		processor string
	}{
		{name: "datadog", processor: "datadog"},
		{name: "otel", processor: "otel"},
		{name: "prometheus", processor: "prometheus"},
		{name: "statsd", processor: "statsd"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, ok := registered[test.processor]; !ok {
				t.Fatalf("registered processors missing %q", test.processor)
			}
		})
	}
}

func registeredProcessors(processors []string) map[string]struct{} {
	registered := make(map[string]struct{}, len(processors))
	for _, processor := range processors {
		registered[processor] = struct{}{}
	}

	return registered
}
