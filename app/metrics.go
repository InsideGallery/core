package app

import (
	"fmt"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
)

// MetricsClientOptions configures metrics initialization without reading environment state.
//
// Deprecated: use app.RunWeb with app.Config.
type MetricsClientOptions struct {
	Config              metrics.Config
	ServiceName         string
	ProcessorRegistry   *metrics.Registry
	HealthState         *profiler.State
	InstallDefault      bool
	RegisterHealthCheck bool
}

// MetricsRuntime owns an initialized metrics client and its close path.
type MetricsRuntime struct {
	client        *metrics.Client
	defaultHandle *metrics.DefaultHandle
}

// NewMetricsClient initializes metrics from explicit options.
func NewMetricsClient(options MetricsClientOptions) (*MetricsRuntime, error) {
	client, err := metrics.NewWithRegistry(options.ProcessorRegistry, options.Config, options.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("metrics init: %w", err)
	}

	runtime := &MetricsRuntime{
		client: client,
	}

	if options.InstallDefault {
		runtime.defaultHandle = metrics.InstallDefault(client)
	}

	if client != nil && options.RegisterHealthCheck {
		healthState := options.HealthState
		if healthState == nil {
			healthState = profiler.DefaultState()
		}

		healthState.AddHealthCheck(client.HealthCheck)
	}

	return runtime, nil
}

// Client returns the initialized metrics client.
func (r *MetricsRuntime) Client() *metrics.Client {
	if r == nil {
		return nil
	}

	return r.client
}

// Close closes the metrics client and restores any scoped default.
func (r *MetricsRuntime) Close() error {
	if r == nil {
		return nil
	}

	if r.defaultHandle != nil {
		return r.defaultHandle.Close()
	}

	return r.client.Close()
}

func newMetricsClient(serviceName string) (*metrics.Client, func() error, error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("metrics config: %w", err)
	}

	runtime, err := NewMetricsClient(MetricsClientOptions{
		Config:              cfg,
		ServiceName:         serviceName,
		InstallDefault:      true,
		RegisterHealthCheck: true,
	})
	if err != nil {
		return nil, nil, err
	}

	return runtime.Client(), func() error {
		return runtime.Close()
	}, nil
}
