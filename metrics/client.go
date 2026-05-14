// Package metrics provides backend-agnostic service instrumentation.
//
// Services record metrics through Client or the Processor interface. Concrete
// exporters live in pkg/metrics/processors/* and register themselves at init,
// following the same plugin pattern used by pkg/fastlog.
package metrics //nolint:revive // intentional: "metrics" is a domain name, not stdlib's runtime/metrics

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
)

// Processor records metrics for one concrete backend.
type Processor interface {
	Close() error
	Count(name string, value int64, tags []string) error
	Gauge(name string, value float64, tags []string) error
	Distribution(name string, value float64, tags []string) error
}

// Factory creates a concrete metrics processor for a service.
type Factory func(Config, string) (Processor, error)

// Client fans metric calls out to configured processors.
type Client struct {
	processors []Processor
	service    string
}

var (
	defaultMu     sync.RWMutex
	defaultClient *Client

	registryMu sync.RWMutex
	registry   = map[string]Factory{}
)

// Register makes a metrics processor available by kind.
func Register(kind string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	registry[strings.ToLower(strings.TrimSpace(kind))] = factory
}

// RegisteredProcessors returns all registered processor names.
func RegisteredProcessors() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// SetDefault stores the process-wide metrics client for service-specific instrumentation.
func SetDefault(c *Client) {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	defaultClient = c
}

// Default returns the process-wide metrics client, or nil when metrics are disabled.
func Default() *Client {
	defaultMu.RLock()
	defer defaultMu.RUnlock()

	return defaultClient
}

// DefaultHandle restores a package-level metrics default and closes its client.
type DefaultHandle struct {
	client   *Client
	previous *Client
	once     sync.Once
	err      error
}

// InstallDefault installs a process-wide metrics default with an explicit close path.
func InstallDefault(c *Client) *DefaultHandle {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	previous := defaultClient
	defaultClient = c

	return &DefaultHandle{
		client:   c,
		previous: previous,
	}
}

// Client returns the installed default client.
func (h *DefaultHandle) Client() *Client {
	if h == nil {
		return nil
	}

	return h.client
}

// Close restores the previous default client and closes the installed client.
func (h *DefaultHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		defaultMu.Lock()
		defaultClient = h.previous
		defaultMu.Unlock()

		h.err = h.client.Close()
	})

	return h.err
}

// New creates a metrics client from configured processors.
// Returns nil if cfg is not enabled.
func New(cfg Config, service string) (*Client, error) {
	kinds := cfg.EnabledProcessors()
	if len(kinds) == 0 {
		return nil, nil
	}

	var errs []error

	processors := make([]Processor, 0, len(kinds))

	for _, kind := range kinds {
		processor, err := newProcessor(kind, cfg, service)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		if processor != nil {
			processors = append(processors, processor)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	if len(processors) == 0 {
		return nil, nil
	}

	c := &Client{processors: processors, service: service}

	slog.Info("Metrics enabled", "processors", kinds, "service", service)

	return c, nil
}

//nolint:ireturn // registry boundary returns the abstraction by design
func newProcessor(kind string, cfg Config, service string) (Processor, error) {
	normalized := strings.ToLower(strings.TrimSpace(kind))

	factory, ok := registeredFactory(normalized)

	if !ok {
		return nil, fmt.Errorf("metrics processor %q is not registered", normalized)
	}

	processor, err := factory(cfg, service)
	if err != nil {
		return nil, fmt.Errorf("metrics processor %q: %w", normalized, err)
	}

	return processor, nil
}

func registeredFactory(kind string) (Factory, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	factory, ok := registry[kind]

	return factory, ok
}

// Close flushes pending metrics and closes all processors.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, processor := range c.processors {
		if processor == nil {
			continue
		}

		if err := processor.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Count records a count metric.
func (c *Client) Count(name string, value int64, tags []string) error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, processor := range c.processors {
		if err := processor.Count(name, value, tags); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMetricErrors("count", name, errs)
}

// Gauge records a gauge metric.
func (c *Client) Gauge(name string, value float64, tags []string) error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, processor := range c.processors {
		if err := processor.Gauge(name, value, tags); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMetricErrors("gauge", name, errs)
}

// Distribution records a distribution metric.
func (c *Client) Distribution(name string, value float64, tags []string) error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, processor := range c.processors {
		if err := processor.Distribution(name, value, tags); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMetricErrors("distribution", name, errs)
}

func wrapMetricErrors(operation, name string, errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("metrics %s %q: %w", operation, name, errors.Join(errs...))
}

// NormalizeTags returns a stable copy of tags suitable for processors.
func NormalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	normalized := append([]string(nil), tags...)
	sort.Strings(normalized)

	return normalized
}

// TagSet returns a stable tag-set strings for processors that cannot model arbitrary labels.
func TagSet(tags []string) string {
	return strings.Join(NormalizeTags(tags), ",")
}
