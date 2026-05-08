package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"sync"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

// WriterFactory creates an output writer and handler options.
type WriterFactory func() (io.Writer, *slog.HandlerOptions, error)

// HandlerFactory creates a slog handler.
type HandlerFactory func() slog.Handler

// Registry owns slog handler factories for explicit application composition.
type Registry struct {
	mu               sync.RWMutex
	writers          map[string]WriterFactory
	handlers         map[string]slog.Handler
	handlerFactories map[string]HandlerFactory
}

var (
	defaultRegistryMu sync.RWMutex    //nolint:gochecknoglobals // protects compatibility registry swaps
	defaultRegistry   = NewRegistry() //nolint:gochecknoglobals // compatibility registry for handler init hooks
)

// DefaultRegistry returns the package-level compatibility registry.
func DefaultRegistry() *Registry {
	defaultRegistryMu.RLock()
	defer defaultRegistryMu.RUnlock()

	return defaultRegistry
}

// NewRegistry creates an isolated slog handler registry.
func NewRegistry() *Registry {
	return &Registry{
		writers:          make(map[string]WriterFactory),
		handlers:         make(map[string]slog.Handler),
		handlerFactories: make(map[string]HandlerFactory),
	}
}

// DefaultRegistryHandle restores a previous package-level handler registry.
type DefaultRegistryHandle struct {
	previous *Registry
	once     sync.Once
}

// InstallDefaultRegistry installs a scoped package-level handler registry.
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

// Close restores the previous package-level handler registry.
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

// RegisterWriter stores a writer factory on this registry.
func (r *Registry) RegisterWriter(kind string, writer WriterFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.writers[kind] = writer
}

// RegisterHandler stores a concrete handler on this registry.
func (r *Registry) RegisterHandler(kind string, handler slog.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[kind] = handler
}

// RegisterHandlerFactory stores a handler factory on this registry.
func (r *Registry) RegisterHandlerFactory(kind string, factory HandlerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlerFactories[kind] = factory
}

// Get resolves a handler from this registry.
func (r *Registry) Get(kind, format string, defaultLogLevel slog.Level) (slog.Handler, error) {
	r.mu.RLock()

	handler, ok := r.handlers[kind]
	if ok {
		r.mu.RUnlock()

		return handler, nil
	}

	factory, ok := r.handlerFactories[kind]
	if ok {
		r.mu.RUnlock()

		handler = factory()
		if handler != nil {
			return handler, nil
		}
	} else {
		r.mu.RUnlock()
	}

	r.mu.RLock()

	h, ok := r.writers[kind]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: kind: %s", ErrNotFoundHandler, kind)
	}

	w, opts, err := h()
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &slog.HandlerOptions{
			Level: defaultLogLevel,
		}
	}

	switch format {
	case FormatText:
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}

	return handler, nil
}

// RegisterWriter stores a writer factory on the package-level compatibility registry.
//
// Deprecated: use Registry.RegisterWriter on an explicit registry.
func RegisterWriter(kind string, writer WriterFactory) {
	DefaultRegistry().RegisterWriter(kind, writer)
}

// RegisterHandler stores a concrete handler on the package-level compatibility registry.
//
// Deprecated: use Registry.RegisterHandler on an explicit registry.
func RegisterHandler(kind string, handler slog.Handler) {
	DefaultRegistry().RegisterHandler(kind, handler)
}

// RegisterHandlerFactory stores a handler factory on the package-level compatibility registry.
//
// Deprecated: use Registry.RegisterHandlerFactory on an explicit registry.
func RegisterHandlerFactory(kind string, factory HandlerFactory) {
	DefaultRegistry().RegisterHandlerFactory(kind, factory)
}

// Get resolves a handler from the package-level compatibility registry.
//
// Deprecated: use Registry.Get on an explicit registry.
func Get(kind, format string, defaultLogLevel slog.Level) (slog.Handler, error) {
	return DefaultRegistry().Get(kind, format, defaultLogLevel)
}
