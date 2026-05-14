# db/redis

Import path: `github.com/InsideGallery/core/db/redis`

Package `redis` provides Redis client configuration, connection ownership helpers, and core-owned
key/value command shapes. New code should use `KeyValueStore` where application-facing code should not
depend directly on the go-redis client.

## Main APIs

- `ConnectionConfig` configures host, port, user, password, and database.
- `GetConnectionConfigFromEnv()` reads `REDIS_*` environment variables.
- `NewRedisClient(config)` creates a `Connection` wrapper around `github.com/redis/go-redis/v9`.
- `ConnectionStore` owns a connection for explicit application composition.
- `KeyValueStore` is the core-owned interface for get, set, and stop operations.
- `GetOptions`, `SetOptions`, `StringResult`, and `CommandResult` describe string key/value commands.
- `Connection.Get` returns `present: false` and no error for missing keys.
- `Connection.Set` writes a value with a TTL.
- `Set`, `Get`, and `Default` are legacy package-level connection helpers.

## Usage

```go
package example

import (
	"context"
	"time"

	"github.com/InsideGallery/core/db/redis"
)

func cacheName(ctx context.Context) (err error) {
	connection := redis.NewRedisClient(&redis.ConnectionConfig{
		Host: "localhost",
		Port: "6379",
	})
	defer func() {
		if closeErr := connection.Stop(); err == nil {
			err = closeErr
		}
	}()

	_, err := connection.SetValue(ctx, redis.SetOptions{
		Key:   "profile:1:name",
		Value: "Ada",
		TTL:   time.Minute,
	})

	return err
}
```

## Configuration And Operations

Environment variables include `REDIS_HOST`, `REDIS_PORT`, `REDIS_USER`, `REDIS_PASS`, and
`REDIS_DATABASE`. `ConnectionStore.Close` calls `Stop` and clears the stored connection. Close Redis
connections during shutdown.
