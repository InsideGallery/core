# Aerospike Entity Mocks

Import path: `github.com/InsideGallery/core/db/aerospike/entity/mocks`

This package contains generated GoMock support for tests that depend on the
legacy `entity.Operations` interface. The generated package name is
`mock_entity`, so consumers usually import it with an alias.

## Main APIs

- `NewMockOperations(ctrl)` creates a mock implementation of `entity.Operations`.
- `MockOperations.EXPECT()` returns the recorder used to declare expected calls.
- Generated expectations cover `Execute`, `Exists`, `Get`, `GetBin`, and `GetNamespace`.

## Usage

```go
package example_test

import (
	"testing"

	mockentity "github.com/InsideGallery/core/db/aerospike/entity/mocks"
	"github.com/golang/mock/gomock"
)

func TestUsesEntityOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operations := mockentity.NewMockOperations(ctrl)
	operations.EXPECT().GetBin("status").Return("active", nil)

	// Pass operations to code that accepts entity.Operations.
}
```

## Operational Notes

This package is for tests only. Do not edit `models.go` by hand; it is generated
from `db/aerospike/entity/operations.go`.
