package app

import (
	"fmt"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/profiler"
)

func newMetricsClient(serviceName string) (*metrics.Client, func() error, error) {
	cfg, err := metrics.GetEnvConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("metrics config: %w", err)
	}

	client, err := metrics.New(cfg, serviceName)
	if err != nil {
		return nil, nil, fmt.Errorf("metrics init: %w", err)
	}

	metrics.SetDefault(client)

	if client != nil {
		profiler.AddHealthCheck(client.HealthCheck)
	}

	return client, func() error {
		defer metrics.SetDefault(nil)

		return client.Close()
	}, nil
}
