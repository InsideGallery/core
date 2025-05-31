package utils

import (
	"log/slog"
	"sync"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestBatchSlice(t *testing.T) {
	size := 102
	maxBatch := 10
	sl := make([]int, 0, maxBatch)
	for i := 0; i < size; i++ {
		sl = append(sl, i)
	}
	var wg sync.WaitGroup
	var c int
	ch := BatchSlice(maxBatch, sl)
	for res := range ch {
		res := res
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("bath slice", "res", res)
		}()
		c++
	}
	wg.Wait()
	testutils.Equal(t, c, 11)
}
