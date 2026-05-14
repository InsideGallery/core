# db/redis/mocks

Import path: `github.com/InsideGallery/core/db/redis/mocks`

Package `mock_redis` contains generated GoMock test doubles for the legacy `redis.Client` interface. It
is generated from `db/redis/interface.go` and is not a production API.

## Main APIs

- `NewMockClient(ctrl)` creates a generated mock client.
- `MockClient.EXPECT()` returns the recorder used to set expected calls.
- The generated mock covers `Get`, `Set`, and `Stop`.

## Usage

```go
package example_test

import (
	"context"
	"testing"
	"time"

	"github.com/InsideGallery/core/db/redis/mocks"
	"github.com/golang/mock/gomock"
)

func TestClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := mock_redis.NewMockClient(ctrl)

	client.EXPECT().
		Set(context.Background(), "key", "value", time.Minute).
		Return(nil)

	if err := client.Set(context.Background(), "key", "value", time.Minute); err != nil {
		t.Fatal(err)
	}
}
```

## Operational Notes

Do not import this package from production code. Regenerate it through the `go:generate` directive in
`db/redis/interface.go` when the legacy interface changes.
