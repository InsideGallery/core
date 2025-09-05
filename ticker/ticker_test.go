package ticker

import (
	"context"
	"testing"
	"time"

	"github.com/InsideGallery/core/memory/registry"
)

var store = registry.NewRegistry[string, string, any]()

type ExampleTicker struct{}

func (t *ExampleTicker) GetID() uint64 {
	return store.NextID()
}

func (t *ExampleTicker) Tick(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.NewTimer(time.Second).C:
	}
}

func TestTicker(_ *testing.T) {
	m := NewTickManager()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	m.Add(NewTickHandler(ctx, 10*time.Millisecond, &ExampleTicker{}))
	m.Add(NewTickHandler(ctx, 10*time.Millisecond, &ExampleTicker{}))
	m.Add(NewTickHandler(ctx, 10*time.Millisecond, &ExampleTicker{}))
	m.Run()
}
