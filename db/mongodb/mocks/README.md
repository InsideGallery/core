# db/mongodb/mocks

Import path: `github.com/InsideGallery/core/db/mongodb/mocks`

Package `mock_mongodb` contains generated GoMock test doubles for the legacy `mongodb.Client` interface.
It is generated from `db/mongodb/interface.go` and is not a production API.

## Main APIs

- `NewMockClient(ctrl)` creates a generated mock client.
- `MockClient.EXPECT()` returns the recorder used to set expected calls.
- The generated mock covers methods from `mongodb.Client`, including find, aggregate, insert, update,
  delete, collection, database, connection, and batch update methods.

## Usage

```go
package example_test

import (
	"context"
	"testing"

	"github.com/InsideGallery/core/db/mongodb/mocks"
	"github.com/golang/mock/gomock"
)

func TestClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock_mongodb.NewMockClient(ctrl)

	client.EXPECT().
		FindOne(context.Background(), "profiles", gomock.Any(), gomock.Any()).
		Return(nil)

	if err := client.FindOne(context.Background(), "profiles", map[string]any{}, map[string]any{}); err != nil {
		t.Fatal(err)
	}
}
```

## Operational Notes

Do not import this package from production code. Regenerate it through the `go:generate` directive in
`db/mongodb/interface.go` when the legacy interface changes.
