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

func TestEventManagerScopedState(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "managers own subscription IDs",
			run: func(t *testing.T) {
				t.Helper()

				first := NewEventManager(defaultWorkersCount)
				second := NewEventManager(defaultWorkersCount)
				handler := EventHandlerFunc(func(context.Context) {})

				firstID := first.Subscribe("event", handler)
				secondID := second.Subscribe("event", handler)

				if firstID != 1 {
					t.Fatalf("first manager id = %d, want 1", firstID)
				}

				if secondID != 1 {
					t.Fatalf("second manager id = %d, want 1", secondID)
				}
			},
		},
		{
			name: "default event manager handle restores previous manager",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultEventManager()
				next := NewEventManager(defaultWorkersCount)
				handle := InstallDefaultEventManager(next)

				if got := GetEventManager(); got != next {
					t.Fatal("default event manager was not installed")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default handle: %v", err)
				}

				if got := DefaultEventManager(); got != previous {
					t.Fatal("default event manager was not restored")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
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
