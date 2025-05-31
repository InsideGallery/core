package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HistogramChart struct {
	histogram metric.Int64Histogram
}

func NewHistogramChart(histogram metric.Int64Histogram) *HistogramChart {
	return &HistogramChart{
		histogram: histogram,
	}
}

func (c *HistogramChart) Execute(ctx context.Context, value int64, subject string, keyValues ...string) error {
	if c == nil {
		return nil
	}

	attributes, err := PrepareAttributes(keyValues...)
	if err != nil {
		return err
	}

	attributes = append(attributes, attribute.String(attributeSubject, subject))
	c.histogram.Record(ctx, value, metric.WithAttributes(attributes...))

	return nil
}
