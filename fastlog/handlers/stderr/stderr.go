package stderr

import (
	"io"
	"log/slog"
	"os"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "stderr"

func init() {
	handlers.DefaultRegistry().RegisterWriter(OutKind, New)
}

// NewFromConfig creates a stderr writer from explicit config.
func NewFromConfig(cfg Config) (io.Writer, *slog.HandlerOptions, error) {
	return os.Stderr, &slog.HandlerOptions{
		Level: cfg.Level,
	}, nil
}

// New creates a stderr writer from environment config.
//
// Deprecated: use NewFromConfig with explicit config ownership.
func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return os.Stderr, nil, nil
	}

	return NewFromConfig(*cfg)
}
