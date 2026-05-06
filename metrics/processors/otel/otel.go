package otel

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/InsideGallery/core/metrics"
)

// ProcessorName is the registration name for the OpenTelemetry metrics processor.
const ProcessorName = "otel"

func init() {
	metrics.Register(ProcessorName, New)
}

type processor struct {
	meter   otelmetric.Meter
	service string

	mu         sync.Mutex
	counters   map[string]otelmetric.Int64Counter
	gauges     map[string]otelmetric.Float64Gauge
	histograms map[string]otelmetric.Float64Histogram
}

// New creates an OpenTelemetry metrics processor using the global meter provider.
//
//nolint:ireturn // processor factory returns registry abstraction
func New(_ metrics.Config, service string) (metrics.Processor, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return &processor{
		meter:      otel.GetMeterProvider().Meter(cfg.MeterName),
		service:    service,
		counters:   make(map[string]otelmetric.Int64Counter),
		gauges:     make(map[string]otelmetric.Float64Gauge),
		histograms: make(map[string]otelmetric.Float64Histogram),
	}, nil
}

func (p *processor) Close() error {
	return nil
}

func (p *processor) Count(name string, value int64, tags []string) error {
	counter, err := p.counter(name)
	if err != nil {
		return err
	}

	counter.Add(context.Background(), value, otelmetric.WithAttributes(p.attributes(tags)...))

	return nil
}

func (p *processor) Gauge(name string, value float64, tags []string) error {
	gauge, err := p.gauge(name)
	if err != nil {
		return err
	}

	gauge.Record(context.Background(), value, otelmetric.WithAttributes(p.attributes(tags)...))

	return nil
}

func (p *processor) Distribution(name string, value float64, tags []string) error {
	histogram, err := p.histogram(name)
	if err != nil {
		return err
	}

	histogram.Record(context.Background(), value, otelmetric.WithAttributes(p.attributes(tags)...))

	return nil
}

//nolint:ireturn // OpenTelemetry instruments are interface types
func (p *processor) counter(name string) (otelmetric.Int64Counter, error) {
	normalized := sanitizeName(name)

	p.mu.Lock()
	defer p.mu.Unlock()

	if counter, ok := p.counters[normalized]; ok {
		return counter, nil
	}

	counter, err := p.meter.Int64Counter(normalized)
	if err != nil {
		return nil, fmt.Errorf("create counter %q: %w", name, err)
	}

	p.counters[normalized] = counter

	return counter, nil
}

//nolint:ireturn // OpenTelemetry instruments are interface types
func (p *processor) gauge(name string) (otelmetric.Float64Gauge, error) {
	normalized := sanitizeName(name)

	p.mu.Lock()
	defer p.mu.Unlock()

	if gauge, ok := p.gauges[normalized]; ok {
		return gauge, nil
	}

	gauge, err := p.meter.Float64Gauge(normalized)
	if err != nil {
		return nil, fmt.Errorf("create gauge %q: %w", name, err)
	}

	p.gauges[normalized] = gauge

	return gauge, nil
}

//nolint:ireturn // OpenTelemetry instruments are interface types
func (p *processor) histogram(name string) (otelmetric.Float64Histogram, error) {
	normalized := sanitizeName(name)

	p.mu.Lock()
	defer p.mu.Unlock()

	if histogram, ok := p.histograms[normalized]; ok {
		return histogram, nil
	}

	histogram, err := p.meter.Float64Histogram(normalized)
	if err != nil {
		return nil, fmt.Errorf("create histogram %q: %w", name, err)
	}

	p.histograms[normalized] = histogram

	return histogram, nil
}

func (p *processor) attributes(tags []string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("service", p.service),
	}

	for _, tag := range metrics.NormalizeTags(tags) {
		key, value, ok := strings.Cut(tag, ":")
		if !ok || key == "" {
			attrs = append(attrs, attribute.String("tag", tag))

			continue
		}

		attrs = append(attrs, attribute.String(sanitizeAttributeKey(key), value))
	}

	return attrs
}

func sanitizeName(name string) string {
	var builder strings.Builder

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_' || r == '-' || r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}

	if builder.Len() == 0 {
		return "metric"
	}

	return builder.String()
}

func sanitizeAttributeKey(key string) string {
	return strings.ReplaceAll(strings.TrimSpace(key), " ", "_")
}
