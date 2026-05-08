package metrics //nolint:revive // package name matches directory/domain usage

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestClientCloseAndHealthCheck(t *testing.T) {
	closeErr := errors.New("close failed")
	healthErr := errors.New("health failed")

	cases := []struct {
		name      string
		client    *Client
		run       func(*Client) error
		wantError string
	}{
		{
			name:   "nil close succeeds",
			client: nil,
			run: func(client *Client) error {
				return client.Close()
			},
		},
		{
			name:   "nil health check succeeds",
			client: nil,
			run: func(client *Client) error {
				return client.HealthCheck()
			},
		},
		{
			name: "close skips nil processors and joins errors",
			client: &Client{processors: []Processor{
				nil,
				&spyProcessor{err: closeErr},
			}},
			run: func(client *Client) error {
				return client.Close()
			},
			wantError: closeErr.Error(),
		},
		{
			name: "health check skips processors without health contract",
			client: &Client{processors: []Processor{
				&spyProcessor{},
			}},
			run: func(client *Client) error {
				return client.HealthCheck()
			},
		},
		{
			name: "health check joins health errors",
			client: &Client{processors: []Processor{
				&healthProcessor{spyProcessor: spyProcessor{}, err: healthErr},
			}},
			run: func(client *Client) error {
				return client.HealthCheck()
			},
			wantError: healthErr.Error(),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := test.run(test.client)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("error = %v, want nil", err)
				}

				return
			}

			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}

func TestClientMetricErrors(t *testing.T) {
	processorErr := errors.New("processor failed")
	client := &Client{processors: []Processor{&spyProcessor{err: processorErr}}}

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "gauge",
			run: func() error {
				return client.Gauge("queue.depth", 1, nil)
			},
		},
		{
			name: "distribution",
			run: func() error {
				return client.Distribution("request.duration", 1, nil)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil || !strings.Contains(err.Error(), processorErr.Error()) {
				t.Fatalf("error = %v, want containing %q", err, processorErr)
			}
		})
	}
}

func TestNilClientMetricMethods(t *testing.T) {
	var client *Client

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "gauge",
			run: func() error {
				return client.Gauge("queue.depth", 1, nil)
			},
		},
		{
			name: "distribution",
			run: func() error {
				return client.Distribution("request.duration", 1, nil)
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err != nil {
				t.Fatalf("nil metric call: %v", err)
			}
		})
	}
}

func TestDefaultHandleNilAndIdempotentClose(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "nil handle",
			run: func(t *testing.T) {
				t.Helper()

				var handle *DefaultHandle
				if handle.Client() != nil {
					t.Fatal("nil handle returned a client")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("nil handle close: %v", err)
				}
			},
		},
		{
			name: "close is idempotent",
			run: func(t *testing.T) {
				t.Helper()

				processor := &spyProcessor{}
				handle := InstallDefault(&Client{processors: []Processor{processor}})
				t.Cleanup(func() {
					SetDefault(nil)
				})

				if handle.Client() == nil {
					t.Fatal("handle client is nil")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close handle: %v", err)
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close handle twice: %v", err)
				}

				closeCalls := 0
				for _, call := range processor.calls {
					if call == "close" {
						closeCalls++
					}
				}

				if closeCalls != 1 {
					t.Fatalf("close calls = %d, want 1", closeCalls)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestRegistryNormalizesRegistration(t *testing.T) {
	registry := NewRegistry()
	registry.Register("  MIXED  ", func(_ Config, _ string) (Processor, error) {
		return nil, nil //nolint:nilnil // nil processor means no-op registration
	})

	client, err := registry.New(Config{Processors: []string{"mixed"}}, "unit")
	if err != nil {
		t.Fatalf("registry new: %v", err)
	}

	if client != nil {
		t.Fatalf("client = %#v, want nil", client)
	}
}

func TestDefaultRegistryAccessors(t *testing.T) {
	const kind = "accessor-test"

	registry := DefaultRegistry()
	registry.Register(kind, func(_ Config, _ string) (Processor, error) {
		return nil, nil //nolint:nilnil // accessor test only verifies registration
	})

	for _, name := range RegisteredProcessors() {
		if name == kind {
			return
		}
	}

	t.Fatalf("registered processors missing %q", kind)
}

func TestRegistryScopedDefaults(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new with explicit registry ignores compatibility registry",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry()
				registry.Register("scoped", func(_ Config, service string) (Processor, error) {
					return &spyProcessor{service: service}, nil
				})

				client, err := NewWithRegistry(registry, Config{Processors: []string{"scoped"}}, "unit")
				if err != nil {
					t.Fatalf("new with registry: %v", err)
				}

				if client == nil || len(client.processors) != 1 {
					t.Fatalf("processors = %#v, want one processor", client)
				}
			},
		},
		{
			name: "default registry handle restores previous registry",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultRegistry()
				next := NewRegistry()
				handle := InstallDefaultRegistry(next)

				if got := DefaultRegistry(); got != next {
					t.Fatal("default registry was not installed")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default registry handle: %v", err)
				}

				if got := DefaultRegistry(); got != previous {
					t.Fatal("default registry was not restored")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestNormalizeTagsEmpty(t *testing.T) {
	if got := NormalizeTags(nil); got != nil {
		t.Fatalf("NormalizeTags(nil) = %#v, want nil", got)
	}
}

func TestBoundaryRecorderMethods(t *testing.T) {
	client := &Client{processors: []Processor{&spyProcessor{}}}
	errorClient := &Client{processors: []Processor{&spyProcessor{err: errors.New("record failed")}}}
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	cases := []struct {
		name      string
		run       func() (RecordResult, error)
		wantKind  string
		wantError bool
	}{
		{
			name: "new from options returns missing processor error",
			run: func() (RecordResult, error) {
				_, err := NewFromOptions(Options{Service: "unit", Processors: []string{"missing-boundary"}})

				return RecordResult{}, err
			},
			wantError: true,
		},
		{
			name: "count metric success",
			run: func() (RecordResult, error) {
				return client.CountMetric(context.Background(), Metric{Name: "requests", Int: 1})
			},
			wantKind: "count",
		},
		{
			name: "gauge metric success",
			run: func() (RecordResult, error) {
				return client.GaugeMetric(context.Background(), Metric{Name: "queue", Float: 1.5})
			},
			wantKind: "gauge",
		},
		{
			name: "distribution metric success",
			run: func() (RecordResult, error) {
				return client.DistributionMetric(context.Background(), Metric{Name: "latency", Float: 2.5})
			},
			wantKind: "distribution",
		},
		{
			name: "count metric context error",
			run: func() (RecordResult, error) {
				return client.CountMetric(canceledCtx, Metric{Name: "requests", Int: 1})
			},
			wantError: true,
		},
		{
			name: "gauge metric context error",
			run: func() (RecordResult, error) {
				return client.GaugeMetric(canceledCtx, Metric{Name: "queue", Float: 1.5})
			},
			wantError: true,
		},
		{
			name: "distribution metric context error",
			run: func() (RecordResult, error) {
				return client.DistributionMetric(canceledCtx, Metric{Name: "latency", Float: 2.5})
			},
			wantError: true,
		},
		{
			name: "count metric processor error",
			run: func() (RecordResult, error) {
				return errorClient.CountMetric(context.Background(), Metric{Name: "requests", Int: 1})
			},
			wantError: true,
		},
		{
			name: "gauge metric processor error",
			run: func() (RecordResult, error) {
				return errorClient.GaugeMetric(context.Background(), Metric{Name: "queue", Float: 1.5})
			},
			wantError: true,
		},
		{
			name: "distribution metric processor error",
			run: func() (RecordResult, error) {
				return errorClient.DistributionMetric(context.Background(), Metric{Name: "latency", Float: 2.5})
			},
			wantError: true,
		},
		{
			name: "nil recorder returns result",
			run: func() (RecordResult, error) {
				var nilClient *Client

				return nilClient.CountMetric(context.Background(), Metric{Name: "noop", Int: 1})
			},
			wantKind: "count",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.run()
			if test.wantError {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("record metric: %v", err)
			}

			if got.Kind != test.wantKind {
				t.Fatalf("kind = %q, want %q", got.Kind, test.wantKind)
			}
		})
	}
}

type healthProcessor struct {
	spyProcessor
	err error
}

func (h *healthProcessor) HealthCheck() error {
	return h.err
}
