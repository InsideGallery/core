package app

import (
	"errors"
	"testing"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
)

func TestNewMetricsClient(t *testing.T) {
	cases := []struct {
		name    string
		options MetricsClientOptions
	}{
		{
			name: "disabled metrics has close path",
			options: MetricsClientOptions{
				Config:      metrics.Config{Processors: []string{"disabled"}},
				ServiceName: "unit",
			},
		},
		{
			name: "disabled metrics can install scoped default",
			options: MetricsClientOptions{
				Config:         metrics.Config{Processors: []string{"disabled"}},
				ServiceName:    "unit",
				InstallDefault: true,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := NewMetricsClient(test.options)
			if err != nil {
				t.Fatalf("new metrics client: %v", err)
			}

			if runtime.Client() != nil {
				t.Fatal("disabled metrics should not create a client")
			}

			if err := runtime.Close(); err != nil {
				t.Fatalf("close metrics runtime: %v", err)
			}
		})
	}
}

func TestNewMetricsClientCompatibility(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "env wrapper returns disabled runtime"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("METRICS_PROCESSORS", "disabled")

			client, closeMetrics, err := newMetricsClient("unit")
			if err != nil {
				t.Fatalf("new metrics client: %v", err)
			}

			if client != nil {
				t.Fatal("disabled metrics should not create a client")
			}

			if err := closeMetrics(); err != nil {
				t.Fatalf("close metrics: %v", err)
			}
		})
	}
}

func TestNewMetricsClientRegistersHealthCheckOnExplicitState(t *testing.T) {
	expectedErr := errors.New("metrics offline")

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "explicit health state owns registered check",
			run: func(t *testing.T) {
				t.Helper()

				registry := metrics.NewRegistry()
				registry.Register("health", func(_ metrics.Config, _ string) (metrics.Processor, error) {
					return healthCheckProcessor{err: expectedErr}, nil
				})

				state := profiler.NewState()
				runtime, err := NewMetricsClient(MetricsClientOptions{
					Config: metrics.Config{
						Processors: []string{"health"},
					},
					ServiceName:         "unit",
					ProcessorRegistry:   registry,
					HealthState:         state,
					RegisterHealthCheck: true,
				})
				if err != nil {
					t.Fatalf("new metrics client: %v", err)
				}
				t.Cleanup(func() {
					if err := runtime.Close(); err != nil {
						t.Fatalf("close metrics runtime: %v", err)
					}
				})

				if err := state.CheckHealth(); !errors.Is(err, expectedErr) {
					t.Fatalf("health check = %v, want %v", err, expectedErr)
				}

				if err := profiler.NewState().CheckHealth(); err != nil {
					t.Fatalf("fresh state health = %v, want nil", err)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

type healthCheckProcessor struct {
	err error
}

func (h healthCheckProcessor) Close() error {
	return nil
}

func (h healthCheckProcessor) Count(string, int64, []string) error {
	return nil
}

func (h healthCheckProcessor) Gauge(string, float64, []string) error {
	return nil
}

func (h healthCheckProcessor) Distribution(string, float64, []string) error {
	return nil
}

func (h healthCheckProcessor) HealthCheck() error {
	return h.err
}
