package ticker

import (
	"context"
	"testing"
	"time"

	"github.com/InsideGallery/core/testutils"
)

func TestExecuteWithDelay(t *testing.T) {
	e := NewExecuteWithDelay()
	e.Start(context.Background(), func(_ context.Context) {}, 10*time.Second)
	testutils.Equal(t, e.IsActive(), true)
	e.Stop()
	testutils.Equal(t, e.IsActive(), false)
}
