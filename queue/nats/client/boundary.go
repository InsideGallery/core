package client

import (
	"context"
	"fmt"
	"time"
)

// ConnectOptions is the core-owned input for creating a NATS connection.
type ConnectOptions struct {
	Addr                 string
	Username             string
	Password             string
	Seed                 string
	DrainTimeout         time.Duration
	MaxReconnects        int
	ReconnectWait        time.Duration
	MaxAckPending        int
	RetryOnFailedConnect bool
	ManualAck            bool
	ConcurrentSize       int
	MaxConcurrentSize    uint64
	ReadTimeout          time.Duration
	IdleTimeout          time.Duration
}

// ConnectionResult is the core-owned result for NATS connection operations.
type ConnectionResult struct {
	Connected bool
}

// CoreClient wraps the legacy NATS client behind core-owned connection operations.
type CoreClient struct {
	client *Client
}

// Connect creates a NATS client from core-owned options.
func Connect(ctx context.Context, options ConnectOptions, logger Logger) (*CoreClient, error) {
	cfg := options.config()

	client, err := ConnectClient(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	return &CoreClient{client: client}, nil
}

// Status reports whether the connection is active.
func (c *CoreClient) Status() ConnectionResult {
	return ConnectionResult{Connected: c.client.Conn().IsConnected()}
}

// Close drains and closes the NATS connection.
func (c *CoreClient) Close() error {
	return c.client.Close()
}

func (o ConnectOptions) config() *Config {
	return &Config{
		Addr:                 o.Addr,
		Username:             o.Username,
		Password:             o.Password,
		Seed:                 o.Seed,
		DrainTimeout:         o.DrainTimeout,
		MaxReconnects:        o.MaxReconnects,
		ReconnectWait:        o.ReconnectWait,
		MaxAckPending:        o.MaxAckPending,
		RetryOnFailedConnect: o.RetryOnFailedConnect,
		ManualAck:            o.ManualAck,
		ConcurrentSize:       o.ConcurrentSize,
		MaxConcurrentSize:    o.MaxConcurrentSize,
		ReadTimeout:          o.ReadTimeout,
		IdleTimeout:          o.IdleTimeout,
	}
}
