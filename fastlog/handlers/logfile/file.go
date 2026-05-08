// Package logfile provides a legacy opt-in slog writer that appends log events to a local file.
//
// New code should select stdout or stderr handlers through fastlog configuration:
//
//	import "github.com/InsideGallery/core/fastlog"
//
// Deprecated: Twelve-Factor applications should use stdout or stderr structured
// event streams and let the runtime route or store logs. This handler remains
// available for compatibility only. If an existing deployment still requires a
// local file, use NewFromConfig so the filename and level are explicit.
package logfile

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const (
	OutKind = "file"
	perm    = 0o666
)

func init() {
	handlers.DefaultRegistry().RegisterWriter(OutKind, New)
}

// NewFromConfig creates a legacy logfile writer from explicit config.
func NewFromConfig(cfg Config) (io.Writer, *slog.HandlerOptions, error) {
	w, err := os.OpenFile(cfg.Name, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)

	return w, &slog.HandlerOptions{
		Level: cfg.Level,
	}, err
}

// New creates a legacy logfile writer from environment config.
//
// Deprecated: use stdout or stderr outputs for Twelve-Factor logging. Use
// NewFromConfig only when local file logging is required for compatibility.
func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	return NewFromConfig(*cfg)
}
