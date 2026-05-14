# db/postgres

Import path: `github.com/InsideGallery/core/db/postgres`

Package `postgres` provides Postgres configuration, sqlx client ownership helpers, and a core-owned SQL
operation boundary. New code should prefer `Database` or `DatabaseClient` when callers do not need direct
access to `sqlx.DB`.

## Main APIs

- `ConnectionConfig` configures host, port, user, password, database, application name, and pool settings.
- `GetConnectionConfigFromEnv()` reads `POSTGRES_*` environment variables.
- `DatabaseOptions` is a core-owned constructor input that accepts `time.Duration` for connection lifetime.
- `NewDatabase`, `NewDatabaseFromOptions`, `DefaultDatabase`, and `WrapDatabase` create `DatabaseClient`
  values.
- `Database` is the core-owned interface for ping, exec, query, query-row, and close operations.
- `Statement` supplies SQL text and arguments.
- `CommandResult` reports rows affected by commands.
- `NewClient`, `ClientStore`, `Set`, `Get`, and `Default` provide legacy sqlx-shaped access.

## Usage

```go
package example

import (
	"context"
	"time"

	"github.com/InsideGallery/core/db/postgres"
)

func insertEvent(ctx context.Context) (err error) {
	database, err := postgres.NewDatabaseFromOptions(postgres.DatabaseOptions{
		Host:            "localhost",
		Port:            "5432",
		User:            "app",
		Password:        "secret",
		Database:        "events",
		ApplicationName: "worker",
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(); err == nil {
			err = closeErr
		}
	}()

	_, err = database.Exec(ctx, postgres.Statement{
		Query: "insert into events (name) values ($1)",
		Args:  []any{"created"},
	})

	return err
}
```

## Configuration And Operations

Environment variables include `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`,
`POSTGRES_DB`, `POSTGRES_APPLICATIONNAME`, `POSTGRES_MAXOPENCONNS`, and `POSTGRES_CONNMAXLIFETIME`.
`ConnectionConfig.GetDSN()` builds a `pgx` DSN with `sslmode=disable` and optional
`fallback_application_name`. `NewClient` opens a handle and configures the pool; callers should use
`Ping` when they need to verify connectivity and must close rows and clients they create.
