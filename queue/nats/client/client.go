package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/metric"
)

type Client struct {
	conn   *nats.Conn
	ctx    context.Context
	cfg    *Config
	logger Logger

	mu    sync.RWMutex
	meter metric.Meter
}

// NewClient creates the legacy NATS SDK-shaped client.
//
// Deprecated: use Connect and CoreClient for new connection code.
func NewClient(
	ctx context.Context,
	conn *nats.Conn,
	cfg *Config,
	logger Logger,
) (*Client, error) {
	return &Client{
		conn:   conn,
		ctx:    ctx,
		cfg:    cfg,
		logger: logger,
	}, conn.Flush()
}

func (c *Client) WithMeter(m metric.Meter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.meter = m
}

func (c *Client) Meter() metric.Meter {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.meter
}

// Conn returns the legacy NATS SDK connection.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// QueueSubscribeSync exposes the legacy NATS subscription type.
func (c *Client) QueueSubscribeSync(subject, queue string) (*nats.Subscription, error) {
	return c.conn.QueueSubscribeSync(subject, queue)
}

func (c *Client) Close() error {
	return c.conn.Drain()
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) Logger() Logger {
	if c.logger == nil {
		return StubLogger{}
	}

	return c.logger
}

func (c *Client) Config() *Config {
	return c.cfg
}

// ConnectClient creates a NATS client from explicit config.
func ConnectClient(ctx context.Context, config *Config, logger Logger) (*Client, error) {
	options, err := config.GetOptionsStrict()
	if err != nil {
		return nil, fmt.Errorf("queue config options: %w", err)
	}

	ncc, err := nats.Connect(config.Addr, options...)
	if err != nil {
		return nil, fmt.Errorf("queue connect err: %w", err)
	}

	return NewClient(ctx, ncc, config, logger)
}

// Default creates a NATS client from environment variables.
//
// Deprecated: use ConnectClient with explicit config or Connect with core-owned options.
func Default(ctx context.Context, logger Logger, prefixes ...string) (*Client, error) {
	config, err := GetNATSConnectionConfigFromEnv(prefixes...)
	if err != nil {
		return nil, fmt.Errorf("error getting queue config: %w", err)
	}

	return ConnectClient(ctx, config, logger)
}
