# db/neo4j

Import path: `github.com/InsideGallery/core/db/neo4j`

Package `neo4j` provides Neo4j driver configuration and a core-owned graph client boundary. New code
should use `Graph` or `GraphClient` so application-facing code does not expose Neo4j SDK driver types.

## Main APIs

- `Options` configures host, credentials, realm, Kerberos ticket, bearer token, and auth type.
- `NewGraphClient(ctx, options)` creates a `GraphClient` and verifies connectivity before returning.
- `Graph` is the core-owned interface for connectivity verification and close operations.
- `Result` reports graph operation status.
- `ConnectionConfig` and `GetConnectionConfigFromEnv()` read `NEO4J_*` environment variables.
- `ConnectionConfig.TokenManager(m)` returns the supplied token manager or creates one from config.
- `TypeBasicAuth`, `TypeKerberosAuth`, and `TypeBearerAuth` select supported auth modes. Other values use
  Neo4j no-auth.
- `Client` and `GetConnection` are legacy SDK-shaped APIs.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/neo4j"
)

func verify(ctx context.Context) (err error) {
	client, err := neo4j.NewGraphClient(ctx, neo4j.Options{
		Host:     "neo4j://127.0.0.1:7687",
		Login:    "neo4j",
		Password: "secret",
		TypeAuth: neo4j.TypeBasicAuth,
	})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := client.Close(ctx); err == nil {
			err = closeErr
		}
	}()

	_, err = client.Verify(ctx)
	return err
}
```

## Configuration And Operations

Environment variables include `NEO4J_LOGIN`, `NEO4J_PASSWORD`, `NEO4J_REALM`, `NEO4J_TICKET`,
`NEO4J_TOKEN`, `NEO4J_HOST`, and `NEO4J_AUTH`. `NewGraphClient` closes the driver if verification fails.
Close the graph client during shutdown.
