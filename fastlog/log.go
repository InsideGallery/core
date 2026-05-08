// Package fastlog configures structured slog handlers.
//
// Environment defaults emit JSON events to stderr. The package registers stdout
// and stderr stream handlers by default:
//
//	import "github.com/InsideGallery/core/fastlog"
//
// New code should prefer NewLoggerWithRegistry, SetupDefaultLogger, and
// InstallDefaultLogger so handler registries and process-wide defaults have
// explicit ownership.
//
// Compatibility: NewLogger and SetupDefaultLog remain available for existing
// package-level wiring. File logging is available only when consumers import
// fastlog/handlers/logfile for legacy compatibility.
package fastlog

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"sync"

	slogmulti "github.com/samber/slog-multi"

	"github.com/InsideGallery/core/fastlog/handlers"
)

var errConfigIsNotSet = errors.New("fastlog config is not set")

// DefaultLoggerHandle restores a previous slog default logger.
type DefaultLoggerHandle struct {
	previous *slog.Logger
	logger   *slog.Logger
	ctx      context.Context
	once     sync.Once
	closeErr error
}

// NewLogger creates a logger from explicit config and the compatibility handler registry.
//
// Deprecated: use SetupDefault for standard bootstrapping or NewLoggerWithRegistry when callers need bespoke
// wiring.
func NewLogger(cfg *Config, m ...slogmulti.Middleware) (*slog.Logger, error) {
	return NewLoggerWithRegistry(cfg, nil, m...)
}

// NewLoggerWithRegistry creates a logger from explicit config and handler registry.
func NewLoggerWithRegistry(
	cfg *Config,
	registry *handlers.Registry,
	m ...slogmulti.Middleware,
) (*slog.Logger, error) {
	if cfg == nil {
		return nil, errConfigIsNotSet
	}

	handler, err := cfg.GetHandlerFromRegistry(registry, m...)
	if err != nil {
		return nil, err
	}

	return slog.New(handler), nil
}

// InstallDefaultLogger installs a process-wide slog default with a restore path.
//
// Use InstallDefaultLogger when callers need bespoke logger wiring. SetupDefault is the standard bootstrap.
func InstallDefaultLogger(logger *slog.Logger) *DefaultLoggerHandle {
	previous := slog.Default()

	if logger != nil {
		slog.SetDefault(logger)
	}

	return &DefaultLoggerHandle{
		previous: previous,
		logger:   logger,
		ctx:      context.Background(),
	}
}

// Close restores the previous process-wide slog default and flushes the installed handler.
func (h *DefaultLoggerHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		slog.SetDefault(h.previous)
		h.closeErr = closeLoggerHandler(h.ctx, h.logger)
	})

	return h.closeErr
}

// SetupDefault creates and installs a process-wide logger from explicit config.
func SetupDefault(
	ctx context.Context,
	cfg *Config,
	m ...slogmulti.Middleware,
) (func() error, error) {
	handle, err := setupDefaultLogger(ctx, cfg, m...)
	if err != nil {
		return nil, err
	}

	return handle.Close, nil
}

// SetupDefaultLogger creates and installs a process-wide logger from explicit config.
func SetupDefaultLogger(cfg *Config, m ...slogmulti.Middleware) (*DefaultLoggerHandle, error) {
	return setupDefaultLogger(context.Background(), cfg, m...)
}

func setupDefaultLogger(
	ctx context.Context,
	cfg *Config,
	m ...slogmulti.Middleware,
) (*DefaultLoggerHandle, error) {
	logger, err := NewLoggerWithRegistry(cfg, nil, m...)
	if err != nil {
		return nil, err
	}

	handle := InstallDefaultLogger(logger)
	handle.ctx = defaultLoggerContext(ctx)

	return handle, nil
}

// SetupDefaultLog reads logger config from environment and installs it as slog default.
//
// Deprecated: use GetConfigFromEnv plus SetupDefault so callers can handle errors and close the logger.
func SetupDefaultLog(m ...slogmulti.Middleware) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := NewLoggerWithRegistry(cfg, nil, m...)
	if err != nil {
		log.Fatal(err)
	}

	slog.SetDefault(logger)
}

func defaultLoggerContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}

func closeLoggerHandler(ctx context.Context, logger *slog.Logger) error {
	if logger == nil {
		return nil
	}

	return closeHandler(ctx, logger.Handler())
}

func closeHandler(ctx context.Context, handler slog.Handler) error {
	if handler == nil {
		return nil
	}

	switch h := handler.(type) {
	case interface{ Close(context.Context) error }:
		return h.Close(ctx)
	case interface{ Shutdown(context.Context) error }:
		return h.Shutdown(ctx)
	case interface{ Flush(context.Context) error }:
		return h.Flush(ctx)
	case interface{ Close() error }:
		return h.Close()
	case interface{ Sync() error }:
		return h.Sync()
	default:
		return nil
	}
}
