package prometheus

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	stdprom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/InsideGallery/core/metrics"
)

// ProcessorName is the registration name for the Prometheus metrics processor.
const ProcessorName = "prometheus"

const (
	contentType = "text/plain; version=0.0.4; charset=utf-8"
)

func init() {
	metrics.DefaultRegistry().Register(ProcessorName, New)
}

type processor struct {
	service string
	cfg     config

	registry *stdprom.Registry
	handler  http.Handler

	mu         sync.Mutex
	counters   map[collectorKey]*stdprom.CounterVec
	gauges     map[collectorKey]*stdprom.GaugeVec
	histograms map[collectorKey]*stdprom.HistogramVec
}

type collectorKey struct {
	name      string
	labelKeys string
}

type labelSet struct {
	names  []string
	values []string
}

var (
	activeMu        sync.RWMutex //nolint:gochecknoglobals // profiler scrape handler reads active processor
	activeProcessor *processor   //nolint:gochecknoglobals // nil means no Prometheus metrics are active
)

// New creates a Prometheus scrape processor.
//
//nolint:ireturn // processor factory returns registry abstraction
func New(_ metrics.Config, service string) (metrics.Processor, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	registry := stdprom.NewRegistry()
	p := &processor{
		service:    service,
		cfg:        cfg,
		registry:   registry,
		handler:    promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}),
		counters:   make(map[collectorKey]*stdprom.CounterVec),
		gauges:     make(map[collectorKey]*stdprom.GaugeVec),
		histograms: make(map[collectorKey]*stdprom.HistogramVec),
	}

	setActiveProcessor(p)

	return p, nil
}

func (p *processor) Close() error {
	clearActiveProcessor(p)

	return nil
}

func (p *processor) Count(name string, value int64, tags []string) error {
	if value < 0 {
		return fmt.Errorf("counter value must not be negative")
	}

	collector, labels, err := p.counter(name, tags)
	if err != nil {
		return err
	}

	collector.WithLabelValues(labels.values...).Add(float64(value))

	return nil
}

func (p *processor) Gauge(name string, value float64, tags []string) error {
	collector, labels, err := p.gauge(name, tags)
	if err != nil {
		return err
	}

	collector.WithLabelValues(labels.values...).Set(value)

	return nil
}

func (p *processor) Distribution(name string, value float64, tags []string) error {
	collector, labels, err := p.histogram(name, tags)
	if err != nil {
		return err
	}

	collector.WithLabelValues(labels.values...).Observe(value)

	return nil
}

// HTTPHandler writes the active Prometheus scrape response.
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	p := currentActiveProcessor()
	if p == nil {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)

		return
	}

	p.handler.ServeHTTP(w, r)
}

func (p *processor) counter(name string, tags []string) (*stdprom.CounterVec, labelSet, error) {
	normalized := sanitizeName(name)
	labels := labelsFromTags(tags)
	key := newCollectorKey(normalized, labels.names)

	p.mu.Lock()
	defer p.mu.Unlock()

	if collector, ok := p.counters[key]; ok {
		return collector, labels, nil
	}

	collector := stdprom.NewCounterVec(stdprom.CounterOpts{
		Name:        normalized,
		Help:        helpText(normalized),
		ConstLabels: stdprom.Labels{"service": p.service},
	}, labels.names)

	if err := p.registry.Register(collector); err != nil {
		return nil, labelSet{}, fmt.Errorf("register counter %q: %w", normalized, err)
	}

	p.counters[key] = collector

	return collector, labels, nil
}

func (p *processor) gauge(name string, tags []string) (*stdprom.GaugeVec, labelSet, error) {
	normalized := sanitizeName(name)
	labels := labelsFromTags(tags)
	key := newCollectorKey(normalized, labels.names)

	p.mu.Lock()
	defer p.mu.Unlock()

	if collector, ok := p.gauges[key]; ok {
		return collector, labels, nil
	}

	collector := stdprom.NewGaugeVec(stdprom.GaugeOpts{
		Name:        normalized,
		Help:        helpText(normalized),
		ConstLabels: stdprom.Labels{"service": p.service},
	}, labels.names)

	if err := p.registry.Register(collector); err != nil {
		return nil, labelSet{}, fmt.Errorf("register gauge %q: %w", normalized, err)
	}

	p.gauges[key] = collector

	return collector, labels, nil
}

func (p *processor) histogram(name string, tags []string) (*stdprom.HistogramVec, labelSet, error) {
	normalized := sanitizeName(name)
	labels := labelsFromTags(tags)
	key := newCollectorKey(normalized, labels.names)

	p.mu.Lock()
	defer p.mu.Unlock()

	if collector, ok := p.histograms[key]; ok {
		return collector, labels, nil
	}

	collector := stdprom.NewHistogramVec(stdprom.HistogramOpts{
		Name:                            normalized,
		Help:                            helpText(normalized),
		ConstLabels:                     stdprom.Labels{"service": p.service},
		Buckets:                         p.cfg.classicBuckets,
		NativeHistogramBucketFactor:     p.cfg.NativeHistogramBucketFactor,
		NativeHistogramZeroThreshold:    p.cfg.NativeHistogramZeroThreshold,
		NativeHistogramMaxBucketNumber:  p.cfg.NativeHistogramMaxBucketNumber,
		NativeHistogramMinResetDuration: p.cfg.NativeHistogramMinResetDuration,
		NativeHistogramMaxZeroThreshold: p.cfg.NativeHistogramMaxZeroThreshold,
	}, labels.names)

	if err := p.registry.Register(collector); err != nil {
		return nil, labelSet{}, fmt.Errorf("register histogram %q: %w", normalized, err)
	}

	p.histograms[key] = collector

	return collector, labels, nil
}

func labelsFromTags(tags []string) labelSet {
	normalizedTags := metrics.NormalizeTags(tags)
	valuesByName := make(map[string]string, len(normalizedTags))
	names := make([]string, 0, len(normalizedTags))

	for _, tag := range normalizedTags {
		name, value, ok := strings.Cut(tag, ":")
		if !ok || name == "" {
			continue
		}

		name = sanitizeLabelName(name)
		if _, ok := valuesByName[name]; ok {
			continue
		}

		valuesByName[name] = value
		names = append(names, name)
	}

	sort.Strings(names)

	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, valuesByName[name])
	}

	return labelSet{names: names, values: values}
}

func newCollectorKey(name string, labelNames []string) collectorKey {
	return collectorKey{
		name:      name,
		labelKeys: strings.Join(labelNames, "\xff"),
	}
}

func helpText(name string) string {
	return "Ptolemy metric " + name + "."
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
		case r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}

	if builder.Len() == 0 {
		return "metric"
	}

	result := builder.String()
	if result[0] >= '0' && result[0] <= '9' {
		return "_" + result
	}

	return result
}

func sanitizeLabelName(name string) string {
	result := sanitizeName(name)
	if result == "metric" {
		return "label"
	}

	return result
}

func setActiveProcessor(p *processor) {
	activeMu.Lock()
	defer activeMu.Unlock()

	activeProcessor = p
}

func clearActiveProcessor(p *processor) {
	activeMu.Lock()
	defer activeMu.Unlock()

	if activeProcessor == p {
		activeProcessor = nil
	}
}

func currentActiveProcessor() *processor {
	activeMu.RLock()
	defer activeMu.RUnlock()

	return activeProcessor
}
