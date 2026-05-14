# db/frogodb

Import path: `github.com/InsideGallery/core/db/frogodb`

Package `frogodb` provides FrogoDB smart-client configuration, connection ownership, and core-owned
record operation types. New code should use `DatabaseClient` or the `Database` interface when callers
should not depend directly on `github.com/FrogoAI/fdb-client/pkg/client`.

## Main APIs

- `ConnectionConfig` configures seeds, timeouts, pool sizes, error-rate limits, and multiplexing.
- `DefaultConnectionConfig(seeds...)` returns the package defaults.
- `GetConnectionConfigFromEnv(prefix)` reads config from `prefix`, defaulting to `FDB` when prefix is empty.
- `NewDatabase(config)` creates a `DatabaseClient`.
- `Database` is the core-owned contract for ping, put, get, delete, count, and close operations.
- `Key`, `PutOptions`, `GetOptions`, `DeleteOptions`, `CountOptions`, and `WriteOptions` describe record
  operations without exposing FrogoDB SDK option types.
- `RecordResult` reports whether a record was found. Missing records return `Found: false` without an
  error.
- `ConnectionRegistry` owns named low-level clients for explicit application composition.
- `NewConnection`, `NewConnectionFromEnv`, `Set`, `Get`, and `Default` expose or manage low-level SDK
  clients for compatibility.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/frogodb"
)

func putRecord(ctx context.Context) (err error) {
	database, err := frogodb.NewDatabase(frogodb.DefaultConnectionConfig("localhost:3000"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(); err == nil {
			err = closeErr
		}
	}()

	_, err = database.PutRecord(ctx, frogodb.PutOptions{
		Key: frogodb.Key{
			Namespace: "app",
			Set:       "profiles",
			Value:     "profile:1",
		},
		Bins: map[string]any{"name": "Ada"},
	})

	return err
}
```

## Configuration And Operations

Environment variables use names such as `FDB_SEEDS`, `FDB_CONNECTION_TIMEOUT`, `FDB_IDLE_TIMEOUT`,
`FDB_POOL_SIZE_PER_NODE`, `FDB_MAX_CONNS_PER_NODE`, and `FDB_MULTIPLEXING`. `CountOptions.AllNodes`
selects the all-node count path. Call `Close` on `DatabaseClient` or `ConnectionRegistry` during
shutdown.
