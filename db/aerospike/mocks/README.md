# Aerospike Mocks

Import path: `github.com/InsideGallery/core/db/aerospike/mocks`

This package contains generated GoMock support for tests that depend on the
legacy Aerospike interfaces from `github.com/InsideGallery/core/db/aerospike`.
The generated package name is `mock_aerospike`, so consumers usually import it
with an alias.

## Main APIs

- `NewMockAerospike(ctrl)` creates a mock for the legacy `aerospike.Aerospike` interface.
- `NewMockNamespace(ctrl)` creates a mock for the legacy `aerospike.Namespace` interface.
- `MockAerospike.EXPECT()` and `MockNamespace.EXPECT()` return recorders for expected calls.
- Generated expectations cover the SDK-shaped methods declared in `interface.go`, including record, batch,
  index, query, UDF, stats, truncate, and existence operations.

## Usage

```go
package example_test

import (
	"testing"

	mockaero "github.com/InsideGallery/core/db/aerospike/mocks"
	"github.com/golang/mock/gomock"
)

func TestUsesNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := mockaero.NewMockNamespace(ctrl)
	namespace.EXPECT().Exists(gomock.Any(), "users", "42").Return(true, nil)

	// Pass namespace to code that accepts aerospike.Namespace.
}
```

## Operational Notes

This package is for tests only. Do not edit `interface.go` by hand; it is
generated from `db/aerospike/interface.go`. For new code that depends on the
core-owned `NamespaceStore` interface, a small local fake at the consumer
boundary is usually simpler than using these legacy SDK-shaped mocks.
