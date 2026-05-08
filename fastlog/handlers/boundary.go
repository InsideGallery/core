package handlers

import (
	"fmt"
	"log/slog"
)

// Options is the core-owned input for resolving a slog handler.
type Options struct {
	Kind            string
	Format          string
	DefaultLogLevel slog.Level
}

// Result is the core-owned result for handler resolution.
type Result struct {
	Kind    string
	Format  string
	Handler slog.Handler
}

// Resolve returns a log handler with metadata that does not expose registered factory internals.
func (r *Registry) Resolve(options Options) (Result, error) {
	handler, err := r.Get(options.Kind, options.Format, options.DefaultLogLevel)
	if err != nil {
		return Result{}, fmt.Errorf("resolve log handler: %w", err)
	}

	return Result{
		Kind:    options.Kind,
		Format:  options.Format,
		Handler: handler,
	}, nil
}

// Resolve returns a log handler with metadata that does not expose registered factory internals.
//
// Deprecated: use Registry.Resolve on an explicit registry.
func Resolve(options Options) (Result, error) {
	return DefaultRegistry().Resolve(options)
}
