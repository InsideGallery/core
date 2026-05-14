# Aerospike Entity

Import path: `github.com/InsideGallery/core/db/aerospike/entity`

This package provides entity-scoped helpers on top of the Aerospike namespace
contracts. New code can use the core-owned `Store`; older code can keep using
the SDK-shaped `Operation` wrapper.

## Main APIs

- `NewStore(namespace, key)` creates a `Store` from an `aerospike.NamespaceStore` and record key.
- `Store.Put`, `Store.Get`, `Store.GetBin`, `Store.Exists`, and `Store.Delete` wrap record operations.
- `RecordStore` is the core-owned interface implemented by `*Store`.
- `BinOptions`, `BinResult`, and `ExistsResult` describe bin and existence results.
- `ErrStoreNotSet` reports a nil store or namespace dependency.
- `NewOperation` creates the legacy operation wrapper around `aerospike.Namespace`.
- `Operation.Execute`, `Get`, `GetBin`, `Exists`, and `GetNamespace` expose the legacy SDK-shaped flow.
- `Operations` is the legacy interface used by the generated mocks package.
- `ErrAttributeNotFound` is returned by legacy `Operation.GetBin` when the record or bin map is missing.

`Store` accepts a nil context by replacing it with `context.Background()`.
Operation errors returned by the namespace store are wrapped with the
`aerospike entity` boundary.

## Usage

```go
package example

import (
	"context"

	aero "github.com/InsideGallery/core/db/aerospike"
	"github.com/InsideGallery/core/db/aerospike/entity"
)

func readUserName(ctx context.Context, namespace aero.NamespaceStore) (any, error) {
	store := entity.NewStore(namespace, aero.Key{Set: "users", Value: "42"})

	result, err := store.GetBin(ctx, entity.BinOptions{Name: "name"})
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return nil, nil
	}

	return result.Value, nil
}
```

## Operational Notes

The legacy `Operation` wrapper builds an Aerospike write policy for `Execute`,
sets `SendKey` from the constructor argument, and applies `Expiration` only when
the supplied expiration is greater than zero.
