# db/gremlin

Import path: `github.com/InsideGallery/core/db/gremlin`

Package `gremlin` provides Gremlin client construction, graph operation helpers, traversal wrappers, and
syntax configuration for Aerospike-style and Neptune-style graph backends. New code should use the
core-owned `VertexStore` or `GraphStore` contracts where possible.

## Main APIs

- `Options` and `NewClient(options)` create a Gremlin `Client` from a URL.
- `GraphStore` supports vertex upsert, edge upsert, vertex counts, value listing, and close operations.
- `UpsertVertexOptions`, `UpsertEdgeOptions`, `CountVerticesOptions`, and `ListValuesOptions` describe
  graph operations without exposing Gremlin SDK traversal values.
- `GraphResult`, `CountResult`, and `ValueListResult` are core-owned result types.
- `ConnectionConfig` and `GetConnectionConfigFromEnv()` read `GREMLIN_URL`.
- `SyntaxConfig`, `GetSyntaxConfigFromEnv()`, `SyntaxState`, and `NewSyntaxState` configure syntax.
- `Cache`, `Operation`, `NewUpsertVertexOp`, `NewUpsertEdgeOp`, `NewCallbackOp`, and `NewDropVertexOp`
  support the legacy operation execution path.
- Package-level `Syntax`, `PropertyID`, `Setup`, and `InstallSyntaxState` remain for compatibility.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/gremlin"
)

func upsertVertex(ctx context.Context) (err error) {
	client, err := gremlin.NewClient(gremlin.Options{URL: "ws://127.0.0.1:8182/gremlin"})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := client.CloseGraph(ctx); err == nil {
			err = closeErr
		}
	}()

	_, err = client.UpsertVertex(ctx, gremlin.UpsertVertexOptions{
		Label:      "person",
		ID:         "person:1",
		Properties: map[string]any{"name": "Ada"},
	})

	return err
}
```

## Configuration And Operations

`GREMLIN_URL` defaults to `ws://127.0.0.1:8182/gremlin`. `GREMLIN_SYNTAX` accepts `aerospike` or
`neptun`; unknown syntax falls back to `aerospike`. The core-owned methods check context cancellation
before building traversals, then wrap operation errors with a Gremlin boundary label. Close the client
during shutdown.
