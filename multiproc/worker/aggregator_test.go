package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestWaiter(t *testing.T) {
	const (
		totalItemsToAdd  = 23
		maxCountPerBatch = 10
		pendingItems     = totalItemsToAdd - 2*maxCountPerBatch
	)

	t.Run("Flushes Count Batches And Drains Pending Items On Close", func(t *testing.T) {
		batches := make(chan int, 3)
		flusherErr := make(chan error, 1)

		agg := NewAggregator[int](
			context.Background(),
			1,
			maxCountPerBatch,
			time.Hour,
			func(items []int) error {
				batches <- len(items)
				return nil
			},
		)

		t.Cleanup(agg.Close)

		go func() {
			flusherErr <- agg.Flusher()
		}()

		addAggregatorItems(t, agg, 0, maxCountPerBatch)
		require.Equal(t, maxCountPerBatch, waitForAggregatorBatch(t, batches))
		require.Equal(t, 0, agg.Count())

		addAggregatorItems(t, agg, maxCountPerBatch, maxCountPerBatch)
		require.Equal(t, maxCountPerBatch, waitForAggregatorBatch(t, batches))
		require.Equal(t, 0, agg.Count())

		addAggregatorItems(t, agg, 2*maxCountPerBatch, pendingItems)
		require.Equal(t, pendingItems, agg.Count())
		requireNoAggregatorBatch(t, batches)

		agg.Close()
		agg.Wait()

		require.Equal(t, pendingItems, waitForAggregatorBatch(t, batches))
		require.NoError(t, waitForAggregatorFlusher(t, flusherErr))
		require.Equal(t, 0, agg.Count())
	})
}

func addAggregatorItems(t *testing.T, agg *Aggregator[int], start, count int) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for i := 0; i < count; i++ {
			agg.Add(start + i)
		}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.FailNow(t, "timed out adding aggregator items")
	}
}

func waitForAggregatorBatch(t *testing.T, batches <-chan int) int {
	t.Helper()

	select {
	case size := <-batches:
		return size
	case <-time.After(time.Second):
		require.FailNow(t, "timed out waiting for aggregator batch")
	}

	return 0
}

func waitForAggregatorFlusher(t *testing.T, flusherErr <-chan error) error {
	t.Helper()

	select {
	case err := <-flusherErr:
		return err
	case <-time.After(time.Second):
		require.FailNow(t, "timed out waiting for aggregator flusher")
	}

	return nil
}

func requireNoAggregatorBatch(t *testing.T, batches <-chan int) {
	t.Helper()

	select {
	case size := <-batches:
		require.Failf(t, "unexpected aggregator batch", "received batch with %d items", size)
	default:
	}
}

func TestAggregator(t *testing.T) {
	t.Run("add values to aggregator", func(t *testing.T) {
		a := NewAggregator(context.TODO(), 1, 100, time.Second, func(_ []any) error {
			return nil
		})
		primID := primitive.NewObjectID()
		a.Add(primID)
		require.Equal(t, a.Count(), 1)
	})

	t.Run("add to aggregator multiple times", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())

		a := NewAggregator(ctx, 1, 100, time.Second, func(_ []any) error {
			return nil
		})
		primID := primitive.NewObjectID()
		a.Add(primID)
		require.Equal(t, a.Count(), 1)

		a.Add(primID)
		require.Equal(t, a.Count(), 2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		err := a.Flusher()
		require.NoError(t, err)
		require.Equal(t, a.Count(), 0)
	})

	t.Run("test lock", func(t *testing.T) {
		a := NewAggregator(context.TODO(), 1, 3, time.Second, func(_ []any) error {
			return nil
		})
		primID := primitive.NewObjectID()

		var counter int32

		go func() {
			for i := 0; i < 4; i++ {
				atomic.AddInt32(&counter, 1)
				a.Add(primID)
			}
		}()

		time.Sleep(time.Millisecond * 10)

		if atomic.LoadInt32(&counter) == 3 {
			err := a.Process()
			require.NoError(t, err)
		}
	})

	t.Run("error processing", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())

		expectedErr := errors.New("mock error")
		a := NewAggregator(ctx, 2, 100, time.Second, func(_ []any) error {
			return expectedErr
		})

		primID := primitive.NewObjectID()
		a.Add(primID)
		require.Equal(t, a.Count(), 1)

		cancel()

		err := a.Flusher()
		require.Equal(t, expectedErr, err)
	})
}
