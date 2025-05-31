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

func (c *Client) Conn() *nats.Conn {
	return c.conn
}

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

func Default(ctx context.Context, logger Logger, prefixes ...string) (*Client, error) {
	config, err := GetNATSConnectionConfigFromEnv(prefixes...)
	if err != nil {
		return nil, fmt.Errorf("error getting queue config: %w", err)
	}

	ncc, err := nats.Connect(config.Addr, config.GetOptions()...)
	if err != nil {
		return nil, fmt.Errorf("queue connect err: %w", err)
	}

	return NewClient(ctx, ncc, config, logger)
}
