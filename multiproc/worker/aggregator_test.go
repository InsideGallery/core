package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAggregator(t *testing.T) {
	t.Run("add values to aggregator", func(t *testing.T) {
		a := NewAggregator(context.TODO(), 100, time.Second, func(_ []any) error {
			return nil
		})
		primID := primitive.NewObjectID()
		a.Add(primID)
		require.Equal(t, a.Count(), 1)
	})

	t.Run("add to aggregator multiple times", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())

		a := NewAggregator(ctx, 100, time.Second, func(_ []any) error {
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
		a := NewAggregator(context.TODO(), 3, time.Second, func(_ []any) error {
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
}
