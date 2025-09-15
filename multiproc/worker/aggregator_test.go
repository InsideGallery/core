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
