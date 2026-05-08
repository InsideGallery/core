package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/InsideGallery/core/testutils"
)

func TestFunctionWithTimeout(t *testing.T) {
	cases := []struct {
		name    string
		timeout time.Duration
		run     func(context.Context) error
		want    error
	}{
		{
			name:    "returns timeout and cancels function context",
			timeout: time.Millisecond,
			run: func(ctx context.Context) error {
				<-ctx.Done()
				time.Sleep(5 * time.Millisecond)

				return nil
			},
			want: ErrFunctionTimeout,
		},
		{
			name:    "returns function result before timeout",
			timeout: time.Second,
			run: func(_ context.Context) error {
				return nil
			},
			want: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			done := make(chan struct{})
			fn := FunctionWithTimeout(test.timeout, func(ctx context.Context) error {
				defer close(done)

				return test.run(ctx)
			})

			testutils.Equal(t, fn(context.Background()), test.want)
			waitFunctionDone(t, done)
		})
	}
}

func waitFunctionDone(t *testing.T, done <-chan struct{}) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("function did not finish")
	}
}

func TestRunSyncMultipleWorkers(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int, 100)

	var count int32

	go func(ch chan int) {
		for i := 0; i < 100; i++ {
			ch <- i
		}

		close(ch)
	}(ch)

	RunSyncMultipleWorkers(ctx, 4, func(_ context.Context) {
		for range ch {
			atomic.AddInt32(&count, 1)
		}
	})

	testutils.Equal(t, atomic.LoadInt32(&count), int32(100))
}

func TestRunAsyncMultipleWorkers(t *testing.T) {
	ctx := context.Background()
	in := make(chan int, 100)

	go func() {
		for i := 0; i < 100; i++ {
			in <- i
		}

		close(in)
	}()

	out := RunAsyncMultipleWorkers(ctx, 4, 100, func(_ context.Context, ch chan<- interface{}) {
		for v := range in {
			ch <- v
		}
	})

	var count int32
	for range out {
		atomic.AddInt32(&count, 1)
	}

	testutils.Equal(t, count, atomic.LoadInt32(&count))
}

func TestFanInOut(t *testing.T) {
	ctx := context.Background()
	in := make(chan interface{}, 100)

	go func() {
		for i := 0; i < 1; i++ {
			in <- i
		}

		close(in)
	}()

	var c int32

	FanInOut(ctx, 4, 100, in, func(_ context.Context, _ interface{}) {
		atomic.AddInt32(&c, 1)
	})
	testutils.Equal(t, atomic.LoadInt32(&c), int32(4))
}
