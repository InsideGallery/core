package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/server/webserver"
)

func TestNewMetricsBoundary(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "disabled")

	client, result, closeMetrics, err := NewMetrics(MetricsOptions{ServiceName: "unit"})
	if err != nil {
		t.Fatalf("new metrics: %v", err)
	}

	if client != nil {
		t.Fatal("client = non-nil, want nil for disabled metrics")
	}

	if result.ServiceName != "unit" {
		t.Fatalf("service name = %q, want unit", result.ServiceName)
	}

	if result.MetricsEnabled {
		t.Fatal("metrics should be disabled")
	}

	if err := closeMetrics(); err != nil {
		t.Fatalf("close metrics: %v", err)
	}
}

func TestWebMainReturnsWhenRouterFails(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "disabled")
	t.Setenv("MONITOR_ADDR", "")

	routerErr := errors.New("router failed")
	called := false

	WebMainWithOptions(context.Background(), WebOptions{
		Port:       "127.0.0.1:0",
		ServerName: "unit",
	}, func(_ context.Context, _ *fiber.App, met *metrics.Client) error {
		called = true

		if met != nil {
			t.Fatal("metrics client = non-nil, want nil")
		}

		return routerErr
	})

	if !called {
		t.Fatal("router was not called")
	}
}

func TestRunWebReturnsRouterError(t *testing.T) {
	routerErr := errors.New("router failed")

	cases := []struct {
		name    string
		options WebOptions
	}{
		{
			name: "router setup failure is returned",
			options: WebOptions{
				Port:       "127.0.0.1:0",
				ServerName: "unit",
				Metrics: MetricsClientOptions{
					Config:      metrics.Config{Processors: []string{"disabled"}},
					ServiceName: "unit",
				},
				InitRouter: func(_ context.Context, _ *fiber.App, _ *metrics.Client) error {
					return routerErr
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := RunWeb(context.Background(), test.options)
			if !errors.Is(err, routerErr) {
				t.Fatalf("err = %v, want router error", err)
			}
		})
	}
}

func TestRunWebReturnsRoutesError(t *testing.T) {
	routesErr := errors.New("routes failed")

	cases := []struct {
		name    string
		options WebOptions
	}{
		{
			name: "route setup failure is returned",
			options: WebOptions{
				Port:       "127.0.0.1:0",
				ServerName: "unit",
				Metrics: MetricsClientOptions{
					Config:      metrics.Config{Processors: []string{"disabled"}},
					ServiceName: "unit",
				},
				InitRoutes: func(_ context.Context, _ webserver.Router) error {
					return routesErr
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := RunWeb(context.Background(), test.options)
			if !errors.Is(err, routesErr) {
				t.Fatalf("err = %v, want routes error", err)
			}
		})
	}
}

func TestWebMainStartsAndStopsWithContext(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "disabled")
	t.Setenv("MONITOR_ADDR", "")
	t.Setenv("DEPLOYMENT_ENVIRONMENT", "test")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)

		WebMain(ctx, "127.0.0.1:0", "unit", func(context.Context, *fiber.App, *metrics.Client) error {
			go func() {
				for {
					if profiler.Ready.Load() {
						cancel()

						return
					}

					if ctx.Err() != nil {
						return
					}

					time.Sleep(time.Millisecond)
				}
			}()

			return nil
		})
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("web main did not stop")
	}
}

func TestNATSMainReturnsWhenConfigFails(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "disabled")
	t.Setenv("MONITOR_ADDR", "")
	t.Setenv("NATS_DRAIN_TIMEOUT", "bad")

	NATSMain(context.Background(), nil)
}

func TestRunNATSReturnsMissingConfigError(t *testing.T) {
	cases := []struct {
		name    string
		options NATSOptions
	}{
		{
			name: "missing explicit nats config is returned",
			options: NATSOptions{
				ServiceName: "unit",
				Metrics: MetricsClientOptions{
					Config:      metrics.Config{Processors: []string{"disabled"}},
					ServiceName: "unit",
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := RunNATS(context.Background(), test.options)
			if err == nil {
				t.Fatal("err = nil, want missing config error")
			}
		})
	}
}

func TestRecoverPanic(_ *testing.T) {
	func() {
		defer recoverPanic("unit panic")

		panic("boom")
	}()
}
