package logstash

import (
	"io"
	"log/slog"
	"net"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "logstash"

func init() {
	handlers.RegisterWriter(OutKind, New)
}

func New() (io.Writer, *slog.HandlerOptions, error) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		return nil, nil, err
	}

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
