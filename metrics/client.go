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

// Counter, Gauge and Observer are resolved metric handles: a caller resolves one
// through HandleProvider at wiring time and records through it afterwards, so
// turning a (name, tags) tuple back into a backend child metric happens once
// instead of on every record. On a data path that records the same bounded set of
// tuples forever, that is the difference between a hashed cache lookup per record
// and an increment.
//
// They are aliases to interface literals rather than defined types, and that is
// load-bearing. A consumer pinned to a release of this module that predates
// HandleProvider cannot name metrics.Counter, but it can declare the identical
// literal locally and assert the capability structurally, because Go matches
// method signatures on type identity and a defined type is never identical to any
// other type. The aliases are therefore what let a consumer use handles when the
// linked version provides them and keep using Count/Gauge/Distribution when it
// does not, without bumping its pin in lockstep with this addition. Do not turn
// them into defined types.
type (
	Counter  = interface{ Add(value int64) }
	Gauge    = interface{ Set(value float64) }
	Observer = interface{ Observe(value float64) }
)

// HandleProvider is the optional capability of resolving a metric handle before
// recording. It is deliberately not part of Processor: a processor that does not
// implement it keeps recording through Count, Gauge and Distribution, so
// implementing it stays opt-in per backend and callers detect support with a type
// assertion. A returned handle must be safe for concurrent use and stays valid
// for the life of the processor.
type HandleProvider interface {
	CounterHandle(name string, tags []string) (Counter, error)
	GaugeHandle(name string, tags []string) (Gauge, error)
	DistributionHandle(name string, tags []string) (Observer, error)
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

// CounterHandle resolves one counter handle for the metric, covering every
// configured processor. See HandleProvider.
//
//nolint:ireturn // handle API returns the abstraction by design
func (c *Client) CounterHandle(name string, tags []string) (Counter, error) {
	if c == nil {
		return discardHandle{}, nil
	}

	handles, err := resolveHandles(c, "counter handle", name, tags,
		HandleProvider.CounterHandle, newProcessorCounter)
	if err != nil {
		return nil, err
	}

	if len(handles) == 1 {
		return handles[0], nil
	}

	return counterFanout(handles), nil
}

// GaugeHandle resolves one gauge handle for the metric, covering every configured
// processor. See HandleProvider.
//
//nolint:ireturn // handle API returns the abstraction by design
func (c *Client) GaugeHandle(name string, tags []string) (Gauge, error) {
	if c == nil {
		return discardHandle{}, nil
	}

	handles, err := resolveHandles(c, "gauge handle", name, tags,
		HandleProvider.GaugeHandle, newProcessorGauge)
	if err != nil {
		return nil, err
	}

	if len(handles) == 1 {
		return handles[0], nil
	}

	return gaugeFanout(handles), nil
}

// DistributionHandle resolves one distribution handle for the metric, covering
// every configured processor. See HandleProvider.
//
//nolint:ireturn // handle API returns the abstraction by design
func (c *Client) DistributionHandle(name string, tags []string) (Observer, error) {
	if c == nil {
		return discardHandle{}, nil
	}

	handles, err := resolveHandles(c, "distribution handle", name, tags,
		HandleProvider.DistributionHandle, newProcessorObserver)
	if err != nil {
		return nil, err
	}

	if len(handles) == 1 {
		return handles[0], nil
	}

	return observerFanout(handles), nil
}

// resolveHandles builds one handle per configured processor: resolved through
// HandleProvider where the processor implements it, and adapted from
// Count/Gauge/Distribution where it does not, so a mixed set of processors records
// through a single handle and no backend has to implement the capability to keep
// working.
//
// A resolution error is returned rather than absorbed: the caller asked for a
// handle and does not have one, and it can still record through Count/Gauge/
// Distribution.
func resolveHandles[T any](
	c *Client,
	operation, name string,
	tags []string,
	resolve func(HandleProvider, string, []string) (T, error),
	adapt func(Processor, string, []string) T,
) ([]T, error) {
	// An adapted handle keeps the tag slice for the life of the process, so it is
	// copied: a caller that assembles its labels in a reused or stack-allocated
	// array must be free to reuse the array after resolving.
	if len(tags) > 0 {
		tags = append(make([]string, 0, len(tags)), tags...)
	}

	var errs []error

	handles := make([]T, 0, len(c.processors))

	for _, processor := range c.processors {
		provider, ok := processor.(HandleProvider)
		if !ok {
			handles = append(handles, adapt(processor, name, tags))

			continue
		}

		handle, err := resolve(provider, name, tags)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		handles = append(handles, handle)
	}

	if len(errs) > 0 {
		return nil, wrapMetricErrors(operation, name, errs)
	}

	return handles, nil
}

// discardHandle is the handle a nil Client resolves. A nil Client records nothing
// (see Count), and a caller must be able to record through a resolved handle
// without a nil check, so the handle discards rather than being nil.
type discardHandle struct{}

func (discardHandle) Add(_ int64) {}

func (discardHandle) Set(_ float64) {}

func (discardHandle) Observe(_ float64) {}

// processorCounter, processorGauge and processorObserver adapt a Processor that
// does not implement HandleProvider into a handle. The per-record error the
// Processor returns is dropped because a handle has no error channel by design —
// it is the same error the Count/Gauge/Distribution path reports, and a caller
// that wants it can record through that path instead.
type processorCounter struct {
	processor Processor
	name      string
	tags      []string
}

//nolint:ireturn // adapter returns the handle abstraction by design
func newProcessorCounter(processor Processor, name string, tags []string) Counter {
	return processorCounter{processor: processor, name: name, tags: tags}
}

func (h processorCounter) Add(value int64) {
	_ = h.processor.Count(h.name, value, h.tags)
}

type processorGauge struct {
	processor Processor
	name      string
	tags      []string
}

//nolint:ireturn // adapter returns the handle abstraction by design
func newProcessorGauge(processor Processor, name string, tags []string) Gauge {
	return processorGauge{processor: processor, name: name, tags: tags}
}

func (h processorGauge) Set(value float64) {
	_ = h.processor.Gauge(h.name, value, h.tags)
}

type processorObserver struct {
	processor Processor
	name      string
	tags      []string
}

//nolint:ireturn // adapter returns the handle abstraction by design
func newProcessorObserver(processor Processor, name string, tags []string) Observer {
	return processorObserver{processor: processor, name: name, tags: tags}
}

func (h processorObserver) Observe(value float64) {
	_ = h.processor.Distribution(h.name, value, h.tags)
}

// counterFanout, gaugeFanout and observerFanout record one value into every
// processor's handle. A single-processor client — the common configuration — gets
// its processor's handle directly instead, so the fan-out costs nothing when
// there is nothing to fan out to.
type counterFanout []Counter

func (f counterFanout) Add(value int64) {
	for _, handle := range f {
		handle.Add(value)
	}
}

type gaugeFanout []Gauge

func (f gaugeFanout) Set(value float64) {
	for _, handle := range f {
		handle.Set(value)
	}
}

type observerFanout []Observer

func (f observerFanout) Observe(value float64) {
	for _, handle := range f {
		handle.Observe(value)
	}
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
