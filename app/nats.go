package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	_ "github.com/InsideGallery/core/fastlog/all" // register supported log handlers
	_ "github.com/InsideGallery/core/metrics/all" // register supported metrics processors

	"github.com/FrogoAI/mq-balancer/subscriber"
	mqdriver "github.com/FrogoAI/mq-balancer/subscriber/driver"
	mqclient "github.com/FrogoAI/mq-balancer/subscriber/driver/client"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
)

// InitSubscriptions is a closure that wires service-specific dependencies (DB, etc.)
// and registers NATS subscriptions. If it returns nil, all setup succeeded.
type InitSubscriptions func(ctx context.Context, sub *subscriber.Subscriber) error

// NATSMain is the complete entrypoint for a NATS worker service.
// It handles: logging → profiler → NATS connect → init closure → signals → wait.
// Reads NATS_ADDR, NATS_CONCURRENT_SIZE, NATS_READ_TIMEOUT from environment.
func NATSMain(name, monitorAddr string, initSubs InitSubscriptions) {
	setupLogging(name)

	ctx := context.Background()

	defer profiler.Monitor(monitorAddr)()

	metricsCfg, err := appMetricsConfig()
	if err != nil {
		slog.Error("Metrics config failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}

	mc, err := metrics.New(metricsCfg, name)
	if err != nil {
		slog.Error("Metrics init failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}

	metrics.SetDefault(mc)

	natsClient, err := mqclient.Default(ctx, slog.Default())
	if err != nil {
		slog.Error("NATS connect failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}

	natsSubscriber := mqdriver.NewNATSSubscriber(natsClient)
	if mc != nil {
		natsSubscriber.WithMeter(mc)
	}

	profiler.AddHealthCheck(func() error {
		conn := natsClient.Conn()
		if !conn.IsConnected() {
			return fmt.Errorf("nats: not connected (status=%v)", conn.Status())
		}

		// Actual round-trip: sends PING, waits for PONG.
		if _, err := conn.RTT(); err != nil {
			return fmt.Errorf("nats: ping failed: %w", err)
		}

		return nil
	})

	slog.Info("NATS connected", "url", natsClient.Conn().ConnectedUrl(), "name", name)

	sub := subscriber.NewSubscriber(natsSubscriber)

	if err := initSubs(ctx, sub); err != nil {
		slog.Error("Init failed", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}

	profiler.Started.Store(true)
	profiler.Ready.Store(true)

	shutdown := func() {
		profiler.Ready.Store(false)
		slog.Info("Shutting down", "service", name)

		if err := sub.Close(); err != nil {
			slog.Error("Shutdown error", "err", err)
		}

		if err := mc.Close(); err != nil {
			slog.Error("Metrics close error", "err", err)
		}
	}

	listener := oslistener.Get()
	listener.Append(syscall.SIGINT, shutdown)
	listener.Append(syscall.SIGTERM, shutdown)
	listener.Append(syscall.SIGQUIT, shutdown)

	oslistener.Start(ctx, listener)

	if err := sub.Wait(); err != nil {
		slog.Error("Worker stopped", "service", name, "err", err)
		os.Exit(1) //nolint:gocritic // intentional
	}
}
