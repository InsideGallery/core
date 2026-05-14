//go:build !metrics_minimal

package all_test

import (
	"testing"

	_ "github.com/InsideGallery/core/metrics/all"

	"github.com/InsideGallery/core/metrics"
)

func TestAllRegistersMetricsProcessors(t *testing.T) {
	registered := registeredProcessors(metrics.RegisteredProcessors())

	for _, processor := range []string{"datadog", "otel", "prometheus", "statsd"} {
		t.Run(processor, func(t *testing.T) {
			if _, ok := registered[processor]; !ok {
				t.Fatalf("registered processors missing %q", processor)
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
