# Aerospike

Import path: `github.com/InsideGallery/core/db/aerospike`

This package provides Aerospike connection helpers, namespace-scoped record
helpers, and small core-owned contracts for callers that should not depend on
Aerospike SDK types at their own boundaries.

## Main APIs

- `ConnectionConfig` and `GetConnectionConfigFromEnv` read Aerospike connection settings.
- `NewConnection` and `NewConnectionFromEnv` create `*aerospike.Client` values.
- `ConnectionRegistry` stores named clients and can close all registered clients.
- `NamespaceStore` is the core-owned record contract for `PutRecord`, `GetRecord`, and `DeleteRecord`.
- `NamespaceInstance` implements `NamespaceStore` and also exposes legacy SDK-shaped namespace methods.
- `Key`, `PutOptions`, `GetOptions`, `DeleteOptions`, `Record`, `RecordResult`, and `Result` are the core-owned
  request/result types.
- `NewValue` and `NewBin` adapt common Go values to Aerospike values and bins.
- `HLLBin`, `MaxIndexBits`, and `MaxAllowedMinhashBits` are shared HLL defaults.
- `ErrConnectionIsNotSet` is returned when a named client is missing from a registry.

The legacy `Aerospike` and `Namespace` interfaces, package-level `Set`, `Get`,
`Default`, and SDK-shaped `NamespaceInstance` methods remain for compatibility.
Prefer `NamespaceStore` and the core-owned option/result types for new call sites.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/aerospike"
)

func saveUser(ctx context.Context) error {
	namespace, err := aerospike.NewNamespaceInstance("app", aerospike.EnvPrefix)
	if err != nil {
		return err
	}

	_, err = namespace.PutRecord(ctx, aerospike.PutOptions{
		Key: aerospike.Key{Set: "users", Value: "42"},
		Bins: map[string]any{
			"name": "Ada",
		},
	})

	return err
}
```

For explicit client lifecycle management, create a client with `NewConnection`
or `NewConnectionFromEnv`, or keep named clients in a `ConnectionRegistry`.
`ConnectionRegistry.Close` closes registered clients and clears the registry.

## Configuration

`GetConnectionConfigFromEnv(prefix)` uppercases the prefix and reads:

- `<PREFIX>_HOST`, default `127.0.0.1`
- `<PREFIX>_PORT`, default `3000`
- `<PREFIX>_HOSTS`, optional host list used instead of host/port when non-empty
- `<PREFIX>_USERNAME` and `<PREFIX>_PASSWORD`, optional internal-auth credentials
- `<PREFIX>_CONNECTION_QUEUE_SIZE`, default `1000`

`EnvPrefix` is `AEROSPIKE`, so the default variables are `AEROSPIKE_HOST`,
`AEROSPIKE_PORT`, and related names. `CreateClientPolicy` enables internal auth
only when both username and password are set.

## Operational Notes

`Setup(DefaultBufferArchitecture())` is called during package initialization to
disable Aerospike buffer architecture flags for int64 compatibility. Tests cover
that `Setup` is idempotent and ignores nil flag dependencies.
