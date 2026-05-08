package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerLifecycle(t *testing.T) {
	cases := []struct {
		name string
		fn   func(context.Context) error
	}{
		{
			name: "nil handler stops cleanly",
		},
		{
			name: "handler receives context",
			fn: func(ctx context.Context) error {
				if ctx == nil {
					return errors.New("context is nil")
				}

				return nil
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			worker := NewWorker(context.Background(), test.fn)
			if worker.ID() == "" {
				t.Fatal("worker id is empty")
			}

			if worker.Context() == nil {
				t.Fatal("worker context is nil")
			}

			if err := worker.Handle(); err != nil {
				t.Fatalf("handle: %v", err)
			}

			if !worker.IsStopped() {
				t.Fatal("worker should be stopped after handle")
			}

			worker.Cancel()
		})
	}
}

func TestWorkersPoolOperations(t *testing.T) {
	pool := NewWorkersPool[int](context.Background())
	done := make(chan struct{})

	pool.Execute(func(context.Context) error {
		close(done)

		return nil
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not run")
	}

	if pool.Size() != 0 {
		t.Fatalf("pool size = %d, want 0", pool.Size())
	}

	for id := range pool.workers.GetMap() {
		if worker, ok := pool.Get(id); !ok || worker.ID() != id {
			t.Fatalf("worker lookup failed for %q", id)
		}

		pool.Remove(id)
	}

	pool.Stop()

	if err := pool.Wait(); err != nil {
		t.Fatalf("wait: %v", err)
	}
}

func TestTemporalWorkerBranches(t *testing.T) {
	pool := NewWorkersPool[int](context.Background())

	t.Run("closed channel stops", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		if err := pool.TemporalWorker(context.Background(), time.Second, nil, ch, nil); err != nil {
			t.Fatalf("temporal worker: %v", err)
		}
	})

	t.Run("timeout callback runs", func(t *testing.T) {
		called := false
		ch := make(chan int)

		if err := pool.TemporalWorker(context.Background(), time.Nanosecond, func() {
			called = true
		}, ch, nil); err != nil {
			t.Fatalf("temporal worker: %v", err)
		}

		if !called {
			t.Fatal("timeout callback was not called")
		}
	})

	t.Run("handler error is logged and worker continues", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 1
		close(ch)

		if err := pool.TemporalWorker(context.Background(), time.Second, nil, ch, func(context.Context, int) error {
			return errors.New("handler failed")
		}); err != nil {
			t.Fatalf("temporal worker: %v", err)
		}
	})
}

func TestPersistentWorkerBranches(t *testing.T) {
	pool := NewWorkersPool[int](context.Background())

	t.Run("closed channel stops", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		if err := pool.PersistentWorker(context.Background(), ch, nil); err != nil {
			t.Fatalf("persistent worker: %v", err)
		}
	})

	t.Run("handler error is logged and worker continues", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 1
		close(ch)

		if err := pool.PersistentWorker(context.Background(), ch, func(context.Context, int) error {
			return errors.New("handler failed")
		}); err != nil {
			t.Fatalf("persistent worker: %v", err)
		}
	})
}
