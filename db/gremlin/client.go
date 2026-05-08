package gremlin

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"

	"github.com/InsideGallery/core/errors"
)

// Client is the legacy Gremlin SDK-shaped client.
//
// Deprecated: use VertexStore and core-owned option/result types for new code.
type Client struct {
	Connection *gremlingo.DriverRemoteConnection
}

// GetConnection creates the legacy Gremlin remote connection.
//
// Deprecated: use NewClient for new code.
func GetConnection(
	cfg *ConnectionConfig,
	configurations ...func(settings *gremlingo.DriverRemoteConnectionSettings),
) (*gremlingo.DriverRemoteConnection, error) {
	return gremlingo.NewDriverRemoteConnection(cfg.URL, configurations...)
}

func ExecIterate(ch <-chan error) error {
	return <-ch
}

func New(
	cfg *ConnectionConfig,
	configurations ...func(settings *gremlingo.DriverRemoteConnectionSettings),
) (*Client, error) {
	conn, err := GetConnection(cfg, configurations...)
	if err != nil {
		return nil, err
	}

	return &Client{
		Connection: conn,
	}, nil
}

func (c *Client) Close() {
	c.Connection.Close()
}

// S returns the legacy Gremlin traversal source.
//
// Deprecated: use core-owned operation methods such as UpsertVertex for new code.
func (c *Client) S() *gremlingo.GraphTraversalSource {
	return gremlingo.Traversal_().WithRemote(c.Connection)
}

func (c *Client) Execute(cache *Cache, ops ...Operation) error {
	var errs []error

	for _, op := range ops {
		err := op.Execute(cache, c.S())
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Combine(errs...)
}
