package datadog

import (
	"fmt"
	"log/slog"

	"github.com/DataDog/datadog-go/v5/statsd"

	"github.com/InsideGallery/core/metrics"
)

// ProcessorName is the registration name for the Datadog metrics processor.
const ProcessorName = "datadog"

func init() {
	metrics.Register(ProcessorName, New)
}

type processor struct {
	client statsd.ClientInterface
}

// New creates a Datadog DogStatsD metrics processor.
//
//nolint:ireturn // processor factory returns registry abstraction
func New(_ metrics.Config, service string) (metrics.Processor, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	if cfg.Addr == "" {
		return nil, fmt.Errorf("address is required")
	}

	client, err := statsd.New(cfg.Addr, statsd.WithNamespace(cfg.namespacePrefix()), statsd.WithTags([]string{
		"service:" + service,
	}))
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", cfg.Addr, err)
	}

	slog.Info("Datadog metrics enabled", "addr", cfg.Addr, "service", service)

	return &processor{client: client}, nil
}

func (p *processor) Close() error {
	return p.client.Close()
}

func (p *processor) Count(name string, value int64, tags []string) error {
	return p.client.Count(name, value, tags, 1)
}

func (p *processor) Gauge(name string, value float64, tags []string) error {
	return p.client.Gauge(name, value, tags, 1)
}

func (p *processor) Distribution(name string, value float64, tags []string) error {
	return p.client.Distribution(name, value, tags, 1)
}
