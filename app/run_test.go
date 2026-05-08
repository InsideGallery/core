package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/metrics"
)

func TestRunWebConfigInstallsSlogDefault(t *testing.T) {
	routerErr := errors.New("router failed")

	cases := []struct {
		name string
	}{
		{name: "router uses installed default logger"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			previous := slog.Default()
			var records atomic.Int64

			registry := handlers.NewRegistry()
			registry.RegisterHandlerFactory("tracked", func() slog.Handler {
				return &trackingLogHandler{records: &records}
			})

			handle := handlers.InstallDefaultRegistry(registry)
			t.Cleanup(func() {
				if err := handle.Close(); err != nil {
					t.Fatalf("close handler registry: %v", err)
				}
			})

			err := RunWeb(context.Background(), trackedWebConfig("tracked", "disabled"), func(
				_ context.Context,
				_ *fiber.App,
				_ *metrics.Client,
			) error {
				slog.Default().Info("router initialized")

				return routerErr
			})
			if !errors.Is(err, routerErr) {
				t.Fatalf("run web error = %v, want router error", err)
			}

			if records.Load() == 0 {
				t.Fatal("installed default logger did not handle router log")
			}

			if slog.Default() != previous {
				t.Fatal("default logger was not restored")
			}
		})
	}
}

func TestRunWebConfigSlogDefaultReturnsConfiguredLogger(t *testing.T) {
	routerErr := errors.New("router failed")
	defaultErr := errors.New("default logger was not configured")

	cases := []struct {
		name string
	}{
		{name: "router captures output through configured default logger"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			previous := slog.Default()
			var buffer bytes.Buffer

			registry := handlers.NewRegistry()
			registry.RegisterHandlerFactory("buffered", func() slog.Handler {
				return slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelInfo})
			})

			handle := handlers.InstallDefaultRegistry(registry)
			t.Cleanup(func() {
				if err := handle.Close(); err != nil {
					t.Fatalf("close handler registry: %v", err)
				}
			})

			err := RunWeb(context.Background(), trackedWebConfig("buffered", "disabled"), func(
				_ context.Context,
				_ *fiber.App,
				_ *metrics.Client,
			) error {
				if slog.Default() == previous {
					return defaultErr
				}

				slog.Default().Info("router initialized with configured logger")

				return routerErr
			})
			if errors.Is(err, defaultErr) {
				t.Fatalf("run web error = %v, want configured slog default", err)
			}

			if !errors.Is(err, routerErr) {
				t.Fatalf("run web error = %v, want router error", err)
			}

			if !strings.Contains(buffer.String(), "router initialized with configured logger") {
				t.Fatalf("log output = %q, want configured logger output", buffer.String())
			}

			if slog.Default() != previous {
				t.Fatal("default logger was not restored")
			}
		})
	}
}

func TestRunWebConfigClosesMetricsBeforeLoggerRestore(t *testing.T) {
	routerErr := errors.New("router failed")

	cases := []struct {
		name string
	}{
		{name: "metrics close observes installed logger"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var records atomic.Int64
			recorder := &closeOrderRecorder{}

			logRegistry := handlers.NewRegistry()
			logRegistry.RegisterHandlerFactory("tracked", func() slog.Handler {
				return &trackingLogHandler{records: &records}
			})

			logHandle := handlers.InstallDefaultRegistry(logRegistry)
			t.Cleanup(func() {
				if err := logHandle.Close(); err != nil {
					t.Fatalf("close handler registry: %v", err)
				}
			})

			metricsRegistry := metrics.NewRegistry()
			metricsRegistry.Register("tracked", func(_ metrics.Config, _ string) (metrics.Processor, error) {
				return &defaultCheckingProcessor{recorder: recorder}, nil
			})

			metricsHandle := metrics.InstallDefaultRegistry(metricsRegistry)
			t.Cleanup(func() {
				if err := metricsHandle.Close(); err != nil {
					t.Fatalf("close metrics registry: %v", err)
				}
			})

			previous := slog.Default()
			err := RunWeb(context.Background(), trackedWebConfig("tracked", "tracked"), func(
				_ context.Context,
				_ *fiber.App,
				_ *metrics.Client,
			) error {
				recorder.setInstalledLogger(slog.Default())

				return routerErr
			})
			if !errors.Is(err, routerErr) {
				t.Fatalf("run web error = %v, want router error", err)
			}

			if got := recorder.events(); len(got) != 1 || got[0] != "metrics-before-logger" {
				t.Fatalf("close events = %v, want metrics before logger restore", got)
			}

			if slog.Default() != previous {
				t.Fatal("default logger was not restored")
			}
		})
	}
}

func TestRunNATSConfigReturnsMissingConfigError(t *testing.T) {
	cases := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "missing nats config",
			cfg: &Config{
				ServiceName: "unit",
				Log: &fastlog.Config{
					Outputs: []string{"nop:json"},
				},
				Metrics: &metrics.Config{
					Processors: []string{"disabled"},
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := RunNATS(context.Background(), test.cfg, nil)
			if err == nil || !strings.Contains(err.Error(), "nats config is not set") {
				t.Fatalf("run nats error = %v, want missing nats config", err)
			}
		})
	}
}

func trackedWebConfig(logOutput string, metricsProcessor string) *Config {
	return &Config{
		ServerName: "unit",
		Port:       "127.0.0.1:0",
		Log: &fastlog.Config{
			Outputs: []string{logOutput + ":json"},
			Level:   slog.LevelInfo,
		},
		Metrics: &metrics.Config{
			Processors: []string{metricsProcessor},
		},
	}
}

type trackingLogHandler struct {
	records *atomic.Int64
}

func (h *trackingLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *trackingLogHandler) Handle(context.Context, slog.Record) error {
	h.records.Add(1)

	return nil
}

func (h *trackingLogHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *trackingLogHandler) WithGroup(string) slog.Handler {
	return h
}

type defaultCheckingProcessor struct {
	recorder *closeOrderRecorder
}

func (p *defaultCheckingProcessor) Close() error {
	if slog.Default() == p.recorder.installedLogger() {
		p.recorder.add("metrics-before-logger")

		return nil
	}

	p.recorder.add("metrics-after-logger")

	return nil
}

func (p *defaultCheckingProcessor) Count(string, int64, []string) error {
	return nil
}

func (p *defaultCheckingProcessor) Gauge(string, float64, []string) error {
	return nil
}

func (p *defaultCheckingProcessor) Distribution(string, float64, []string) error {
	return nil
}

type closeOrderRecorder struct {
	mu        sync.Mutex
	installed *slog.Logger
	names     []string
}

func (r *closeOrderRecorder) setInstalledLogger(logger *slog.Logger) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.installed = logger
}

func (r *closeOrderRecorder) installedLogger() *slog.Logger {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.installed
}

func (r *closeOrderRecorder) add(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.names = append(r.names, name)
}

func (r *closeOrderRecorder) events() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.names...)
}
