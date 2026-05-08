package metrics

import (
	"context"
	"fmt"
)

// Options is the core-owned configuration for constructing a metrics client.
type Options struct {
	Service    string
	Processors []string
	Registry   *Registry
}

// Metric is the core-owned input for recording a metric value.
type Metric struct {
	Name  string
	Tags  []string
	Int   int64
	Float float64
}

// RecordResult describes a completed metric recording operation.
type RecordResult struct {
	Kind           string
	Name           string
	ProcessorCount int
}

// Recorder is the core-owned metrics boundary for application code.
type Recorder interface {
	CountMetric(ctx context.Context, metric Metric) (RecordResult, error)
	GaugeMetric(ctx context.Context, metric Metric) (RecordResult, error)
	DistributionMetric(ctx context.Context, metric Metric) (RecordResult, error)
	Close() error
	HealthCheck() error
}

// NewFromOptions creates a metrics client from core-owned options.
func NewFromOptions(options Options) (*Client, error) {
	client, err := NewWithRegistry(options.Registry, Config{Processors: options.Processors}, options.Service)
	if err != nil {
		return nil, fmt.Errorf("metrics options: %w", err)
	}

	return client, nil
}

// CountMetric records a count metric through the core-owned Recorder contract.
func (c *Client) CountMetric(ctx context.Context, metric Metric) (RecordResult, error) {
	if err := ctx.Err(); err != nil {
		return RecordResult{}, fmt.Errorf("metrics count %q: %w", metric.Name, err)
	}

	if err := c.Count(metric.Name, metric.Int, metric.Tags); err != nil {
		return RecordResult{}, err
	}

	return c.recordResult("count", metric.Name), nil
}

// GaugeMetric records a gauge metric through the core-owned Recorder contract.
func (c *Client) GaugeMetric(ctx context.Context, metric Metric) (RecordResult, error) {
	if err := ctx.Err(); err != nil {
		return RecordResult{}, fmt.Errorf("metrics gauge %q: %w", metric.Name, err)
	}

	if err := c.Gauge(metric.Name, metric.Float, metric.Tags); err != nil {
		return RecordResult{}, err
	}

	return c.recordResult("gauge", metric.Name), nil
}

// Distribution records a distribution metric through the core-owned Recorder contract.
func (c *Client) DistributionMetric(ctx context.Context, metric Metric) (RecordResult, error) {
	if err := ctx.Err(); err != nil {
		return RecordResult{}, fmt.Errorf("metrics distribution %q: %w", metric.Name, err)
	}

	if err := c.Distribution(metric.Name, metric.Float, metric.Tags); err != nil {
		return RecordResult{}, err
	}

	return c.recordResult("distribution", metric.Name), nil
}

func (c *Client) recordResult(kind string, name string) RecordResult {
	if c == nil {
		return RecordResult{Kind: kind, Name: name}
	}

	return RecordResult{
		Kind:           kind,
		Name:           name,
		ProcessorCount: len(c.processors),
	}
}
