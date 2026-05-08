package logstash

import (
	"io"
	"log/slog"
	"net"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "logstash"

func init() {
	handlers.DefaultRegistry().RegisterWriter(OutKind, New)
}

// NewFromConfig creates a Logstash writer from explicit config.
func NewFromConfig(cfg Config) (io.Writer, *slog.HandlerOptions, error) {
	// to test logstash, execute ncat -l 4242 -k
	addr, err := net.ResolveTCPAddr(cfg.Network, cfg.Host)
	if err != nil {
		return nil, nil, err
	}

	w, err := net.DialTCP(cfg.Network, nil, addr)

	return w, &slog.HandlerOptions{
		Level: cfg.Level,
	}, err
}

// New creates a Logstash writer from environment config.
//
// Deprecated: use NewFromConfig with explicit config ownership.
func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

	return NewFromConfig(*cfg)
}
