package statsd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/InsideGallery/core/metrics"
)

// ProcessorName is the registration name for the StatsD metrics processor.
const ProcessorName = "statsd"

func init() {
	metrics.Register(ProcessorName, New)
}

type processor struct {
	conn      net.Conn
	namespace string
	mu        sync.Mutex
}

// New creates a plain UDP StatsD metrics processor.
//
//nolint:ireturn // processor factory returns registry abstraction
func New(_ metrics.Config, _ string) (metrics.Processor, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	if cfg.Addr == "" {
		return nil, fmt.Errorf("address is required")
	}

	conn, err := (&net.Dialer{}).DialContext(context.Background(), "udp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", cfg.Addr, err)
	}

	return &processor{
		conn:      conn,
		namespace: cfg.namespacePrefix(),
	}, nil
}

func (p *processor) Close() error {
	return p.conn.Close()
}

func (p *processor) Count(name string, value int64, _ []string) error {
	return p.write(name, strconv.FormatInt(value, 10), "c")
}

func (p *processor) Gauge(name string, value float64, _ []string) error {
	return p.write(name, strconv.FormatFloat(value, 'f', -1, 64), "g")
}

func (p *processor) Distribution(name string, value float64, _ []string) error {
	return p.write(name, strconv.FormatFloat(value, 'f', -1, 64), "ms")
}

func (p *processor) write(name, value, kind string) error {
	line := p.namespace + sanitizeName(name) + ":" + value + "|" + kind

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := p.conn.Write([]byte(line)); err != nil {
		return fmt.Errorf("write statsd packet: %w", err)
	}

	return nil
}

func sanitizeName(name string) string {
	var builder strings.Builder

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_' || r == '-' || r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}

	if builder.Len() == 0 {
		return "metric"
	}

	return builder.String()
}
