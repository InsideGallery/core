package ticker

import (
	"context"
	"testing"
	"time"
)

type testTickHandler struct {
	id    uint64
	calls chan struct{}
}

func (h *testTickHandler) Tick(context.Context) {
	h.calls <- struct{}{}
}

func (h *testTickHandler) ID() uint64 {
	return h.GetID()
}

func (h *testTickHandler) GetID() uint64 {
	return h.id
}

func TestTickCounter(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "tick increments and reset clears counter",
			run: func(t *testing.T) {
				t.Helper()

				Reset()
				Tick(context.Background())
				Tick(context.Background())

				if got := Get(); got != 2 {
					t.Fatalf("counter = %d, want 2", got)
				}

				Reset()
				if got := Get(); got != 0 {
					t.Fatalf("counter after reset = %d, want 0", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestTickManagerAdditional(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "stop cancels handlers",
			run: func(t *testing.T) {
				t.Helper()

				manager := NewTickManager()
				handler := &testTickHandler{id: 1, calls: make(chan struct{}, 1)}
				tickHandler := NewTickHandler(context.Background(), time.Hour, handler)
				manager.Add(tickHandler)

				manager.Stop()

				select {
				case <-tickHandler.ctx.Done():
				case <-time.After(time.Second):
					t.Fatal("handler context was not cancelled")
				}
			},
		},
		{
			name: "count ticks in progress starts at zero",
			run: func(t *testing.T) {
				t.Helper()

				manager := NewTickManager()
				if got := manager.CountTicksInProgress(); got != 0 {
					t.Fatalf("ticks in progress = %d, want 0", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestExecuteWithDelayAdditional(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "start ignores active timer and executes once",
			run: func(t *testing.T) {
				t.Helper()

				executor := NewExecuteWithDelay()
				calls := make(chan struct{}, 2)
				handler := func(context.Context) {
					calls <- struct{}{}
				}

				executor.Start(context.Background(), handler, 10*time.Millisecond)
				executor.Start(context.Background(), handler, 10*time.Millisecond)

				select {
				case <-calls:
				case <-time.After(time.Second):
					t.Fatal("handler was not called")
				}

				select {
				case <-calls:
					t.Fatal("handler should only be called once")
				default:
				}

				if executor.IsActive() {
					t.Fatal("executor should be inactive after handler runs")
				}
			},
		},
		{
			name: "stop inactive timer returns false",
			run: func(t *testing.T) {
				t.Helper()

				if NewExecuteWithDelay().Stop() {
					t.Fatal("stop should return false for inactive executor")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
