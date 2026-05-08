// Package metrics provides backend-agnostic service instrumentation.
//
// New code should create clients from explicit registries and pass the resulting
// Recorder or *Client through application composition:
//
//	import "github.com/InsideGallery/core/metrics"
//
//	registry := metrics.NewRegistry()
//	client, err := metrics.NewFromOptions(metrics.Options{Service: "api", Registry: registry})
//
// Prefer Recorder, Metric, RecordResult, NewRegistry, NewWithRegistry, and
// InstallDefault when a scoped compatibility default is still needed.
//
// Compatibility: Register, RegisteredProcessors, SetDefault, Default, and New
// remain available for existing package-level wiring. New code should avoid
// hidden process state and pass clients explicitly.
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

// HealthChecker is implemented by processors that can verify exporter health.
type HealthChecker interface {
	HealthCheck() error
}

// Client fans metric calls out to configured processors.
type Client struct {
	processors []Processor
	service    string
}

// Registry owns metrics processor factories for explicit application composition.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

var (
	defaultMu     sync.RWMutex //nolint:gochecknoglobals // process-wide service metrics client
	defaultClient *Client      //nolint:gochecknoglobals // nil means metrics are disabled

	defaultRegistryMu sync.RWMutex    //nolint:gochecknoglobals // protects compatibility registry swaps
	defaultRegistry   = NewRegistry() //nolint:gochecknoglobals // compatibility registry for processor init hooks
)

// DefaultRegistry returns the package-level compatibility processor registry.
func DefaultRegistry() *Registry {
	defaultRegistryMu.RLock()
	defer defaultRegistryMu.RUnlock()

	return defaultRegistry
}

// NewRegistry creates an isolated metrics processor registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// DefaultRegistryHandle restores a previous package-level processor registry.
type DefaultRegistryHandle struct {
	previous *Registry
	once     sync.Once
}

// InstallDefaultRegistry installs a scoped package-level processor registry.
func InstallDefaultRegistry(registry *Registry) *DefaultRegistryHandle {
	defaultRegistryMu.Lock()
	defer defaultRegistryMu.Unlock()

	if registry == nil {
		registry = NewRegistry()
	}

	previous := defaultRegistry
	defaultRegistry = registry

	return &DefaultRegistryHandle{
		previous: previous,
	}
}

// Close restores the previous package-level processor registry.
func (h *DefaultRegistryHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		defaultRegistryMu.Lock()
		defaultRegistry = h.previous
		defaultRegistryMu.Unlock()
	})

	return nil
}

// Register makes a metrics processor available by kind on this registry.
func (r *Registry) Register(kind string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories[normalizeKind(kind)] = factory
}

// RegisteredProcessors returns all processor names registered on this registry.
func (r *Registry) RegisteredProcessors() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// New creates a metrics client from this registry and configured processors.
// Returns nil if cfg is not enabled.
func (r *Registry) New(cfg Config, service string) (*Client, error) {
	kinds := cfg.EnabledProcessors()
	if len(kinds) == 0 {
		return nil, nil //nolint:nilnil // nil means disabled
	}

	var errs []error

	processors := make([]Processor, 0, len(kinds))

	for _, kind := range kinds {
		processor, err := r.newProcessor(kind, cfg, service)
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
		return nil, nil //nolint:nilnil // all configured processors resolved to no-op
	}

	c := &Client{processors: processors, service: service}

	slog.Default().Info("Metrics enabled", "processors", kinds, "service", service)

	return c, nil
}

//nolint:ireturn // registry boundary returns the abstraction by design
func (r *Registry) newProcessor(kind string, cfg Config, service string) (Processor, error) {
	normalized := normalizeKind(kind)

	factory, ok := r.registeredFactory(normalized)
	if !ok {
		return nil, fmt.Errorf("metrics processor %q is not registered", normalized)
	}

	processor, err := factory(cfg, service)
	if err != nil {
		return nil, fmt.Errorf("metrics processor %q: %w", normalized, err)
	}

	return processor, nil
}

func (r *Registry) registeredFactory(kind string) (Factory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[kind]

	return factory, ok
}

// Register makes a metrics processor available by kind.
//
// Deprecated: use NewRegistry and register processors on the explicit registry.
func Register(kind string, factory Factory) {
	DefaultRegistry().Register(kind, factory)
}

// RegisteredProcessors returns all registered processor names.
//
// Deprecated: use Registry.RegisteredProcessors on an explicit registry.
func RegisteredProcessors() []string {
	return DefaultRegistry().RegisteredProcessors()
}

// SetDefault stores the process-wide metrics client for service-specific instrumentation.
//
// Deprecated: pass *Client explicitly or use InstallDefault for a scoped compatibility default.
func SetDefault(c *Client) {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	defaultClient = c
}

// Default returns the process-wide metrics client, or nil when metrics are disabled.
//
// Deprecated: pass *Client explicitly instead of reading package-level state.
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
//
// Deprecated: use NewWithRegistry or Registry.New with an explicit registry.
func New(cfg Config, service string) (*Client, error) {
	return NewWithRegistry(nil, cfg, service)
}

// NewWithRegistry creates a metrics client from an explicit processor registry.
func NewWithRegistry(registry *Registry, cfg Config, service string) (*Client, error) {
	if registry == nil {
		registry = DefaultRegistry()
	}

	return registry.New(cfg, service)
}

func normalizeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
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

// HealthCheck verifies all processors that expose a health check.
func (c *Client) HealthCheck() error {
	if c == nil {
		return nil
	}

	var errs []error

	for _, processor := range c.processors {
		checker, ok := processor.(HealthChecker)
		if !ok {
			continue
		}

		if err := checker.HealthCheck(); err != nil {
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

// TagSet returns a stable tag-set string for processors that cannot model arbitrary labels.
func TagSet(tags []string) string {
	return strings.Join(NormalizeTags(tags), ",")
}
