package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/oslistener"
	"github.com/InsideGallery/core/profiler"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

// LoggerOptions configures bootstrap logging without forcing environment reads.
//
// Deprecated: use app.RunWeb with app.Config.
type LoggerOptions struct {
	// Deprecated: configure logging with Config.Log and emit logs through slog.Default().
	Logger *slog.Logger
	// Deprecated: configure logging with Config.Log and emit logs through slog.Default().
	Config *fastlog.Config
	// Deprecated: configure logging with Config.Log and emit logs through slog.Default().
	HandlerRegistry *handlers.Registry
	// Deprecated: app.RunWeb and app.RunNATS install slog.Default() from Config.Log.
	InstallDefault bool
}

// RuntimeEnvironmentOptions holds environment-derived runtime toggles.
//
// Deprecated: use app.RunWeb with app.Config.
type RuntimeEnvironmentOptions struct {
	EnablePrintRoutes bool
}

// WebOptions is the core-owned input for web bootstrap helpers.
//
// Deprecated: use app.RunWeb with app.Config.
type WebOptions struct {
	Port       string
	ServerName string
	// Deprecated: configure logging with Config.Log and emit logs through slog.Default().
	Logger           LoggerOptions
	Metrics          MetricsClientOptions
	ProfilerState    *profiler.State
	MonitorAddr      string
	ShutdownListener *oslistener.SignalListener
	ShutdownTimeout  time.Duration
	Environment      RuntimeEnvironmentOptions
	InitRoutes       InitRoutes
	InitRouter       InitRouter
}

// MetricsOptions is the core-owned input for metrics bootstrap helpers.
//
// Deprecated: use app.RunWeb with app.Config.
type MetricsOptions struct {
	ServiceName string
}

// NATSOptions is the core-owned input for NATS bootstrap helpers.
//
// Deprecated: use app.RunNATS with app.Config.
type NATSOptions struct {
	ServiceName string
	// Deprecated: configure logging with Config.Log and emit logs through slog.Default().
	Logger            LoggerOptions
	Metrics           MetricsClientOptions
	ProfilerState     *profiler.State
	MonitorAddr       string
	ShutdownListener  *oslistener.SignalListener
	NATSConfig        *client.Config
	NATSClient        *client.Client
	Subscriber        *subscriber.Subscriber
	Middleware        *middleware.Middleware
	InitSubscriptions InitSubscriptions
	CloseNATSClient   bool
}

// BootstrapResult describes initialized application support services.
//
// Deprecated: use app.RunWeb with app.Config.
type BootstrapResult struct {
	ServiceName    string
	MetricsEnabled bool
}

// WebMainWithOptions starts the web bootstrap flow with core-owned options.
//
// Deprecated: use app.RunWeb with app.Config.
func WebMainWithOptions(ctx context.Context, options WebOptions, initRouter InitRouter) {
	if initRouter != nil {
		options.InitRouter = initRouter
	}

	if err := RunWeb(ctx, options); err != nil {
		slog.Default().Error("run web bootstrap", "err", err)
	}
}

// NewMetrics initializes metrics with a core-owned options and result contract.
func NewMetrics(options MetricsOptions) (*metrics.Client, BootstrapResult, func() error, error) {
	client, closeMetrics, err := newMetricsClient(options.ServiceName)
	if err != nil {
		return nil, BootstrapResult{}, nil, fmt.Errorf("app metrics: %w", err)
	}

	return client, BootstrapResult{
		ServiceName:    options.ServiceName,
		MetricsEnabled: client != nil,
	}, closeMetrics, nil
}
