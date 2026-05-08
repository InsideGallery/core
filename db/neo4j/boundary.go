// Package neo4j provides Neo4j graph client helpers.
//
// New code should use the core-owned graph boundary:
//
//	import "github.com/InsideGallery/core/db/neo4j"
//
//	client, err := neo4j.NewGraphClient(ctx, neo4j.Options{Host: "neo4j://127.0.0.1:7687"})
//
// Prefer Graph, Options, and Result in application-facing code so the Neo4j SDK
// driver type stays inside this adapter package.
//
// Compatibility: Client and GetConnection remain available for existing
// SDK-shaped callers. Prefer NewGraphClient for new integrations.
package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	coreerrors "github.com/InsideGallery/core/errors"
)

// Options is the core-owned input for creating a Neo4j graph client.
type Options struct {
	Host     string
	Login    string
	Password string
	Realm    string
	Ticket   string
	Token    string
	TypeAuth string
}

// Result is the core-owned result for Neo4j graph client operations.
type Result struct {
	Connected bool
}

// Graph is the core-owned Neo4j contract for new consumers.
type Graph interface {
	Verify(ctx context.Context) (Result, error)
	Close(ctx context.Context) error
}

// GraphClient wraps the Neo4j driver behind core-owned operation inputs and results.
type GraphClient struct {
	driver neo4j.DriverWithContext
}

// NewGraphClient creates a Neo4j graph client from core-owned options.
func NewGraphClient(ctx context.Context, options Options) (*GraphClient, error) {
	cfg := options.connectionConfig()

	driver, err := neo4j.NewDriverWithContext(cfg.Host, cfg.TokenManager(nil))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver: %w", err)
	}

	client := &GraphClient{driver: driver}
	if _, err = client.Verify(ctx); err != nil {
		closeErr := client.Close(ctx)
		if closeErr != nil {
			err = fmt.Errorf("%w: %w", err, closeErr)
		}

		return nil, err
	}

	return client, nil
}

// Verify verifies Neo4j connectivity.
func (c *GraphClient) Verify(ctx context.Context) (Result, error) {
	if err := c.driver.VerifyConnectivity(ctx); err != nil {
		return Result{}, coreerrors.WrapBoundary("neo4j", "verify connectivity", err)
	}

	return Result{Connected: true}, nil
}

// Close closes the Neo4j graph client.
func (c *GraphClient) Close(ctx context.Context) error {
	if err := c.driver.Close(ctx); err != nil {
		return coreerrors.WrapBoundary("neo4j", "close", err)
	}

	return nil
}

func (o Options) connectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Login:    o.Login,
		Password: o.Password,
		Realm:    o.Realm,
		Ticket:   o.Ticket,
		Token:    o.Token,
		Host:     o.Host,
		TypeAuth: o.TypeAuth,
	}
}
