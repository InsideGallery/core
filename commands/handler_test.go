package commands

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestHandler(t *testing.T) {
	var calls uint32

	ev := EventHandlerFunc(func(context.Context) {
		atomic.AddUint32(&calls, 1)
	})
	eventManager := NewEventManager(defaultWorkersCount)
	eventManager.Subscribe("event1", ev)
	id := eventManager.Subscribe("event1", ev)
	eventManager.Subscribe("event1", ev)
	eventManager.Subscribe("event2", ev)

	testutils.Equal(t, len(eventManager.GetHandlers("event1")), 3)

	eventManager.Call(context.Background(), "event1")
	testutils.Equal(t, atomic.LoadUint32(&calls), uint32(3))
	eventManager.Call(context.Background(), "event2")
	testutils.Equal(t, atomic.LoadUint32(&calls), uint32(4))

	eventManager.Unsubscribe("event1", id)
	testutils.Equal(t, len(eventManager.GetHandlers("event1")), 2)
}

/*
BenchmarkHandler 500000	      2648 ns/op
*/

func BenchmarkHandler(b *testing.B) {
	eventManager := NewEventManager(defaultWorkersCount)
	eventManager.Subscribe("event1", EventHandlerFunc(func(context.Context) {}))
	eventManager.Subscribe("event1", EventHandlerFunc(func(context.Context) {}))
	eventManager.Subscribe("event2", EventHandlerFunc(func(context.Context) {}))

	for i := 0; i < b.N; i++ {
		eventManager.Call(context.Background(), "event1")
	}

	eventManager.Unsubscribe("event1", 2)
}
