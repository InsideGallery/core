package neo4j

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Client struct {
	neo4j.DriverWithContext
	ctx context.Context
}

func GetConnection(
	ctx context.Context,
	cfg *ConnectionConfig,
) (*Client, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.Host,
		cfg.TokenManager(nil),
	)
	if err != nil {
		return nil, err
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{ctx: ctx, DriverWithContext: driver}, nil
}

func (c *Client) Close() error {
	return c.DriverWithContext.Close(c.ctx)
}
