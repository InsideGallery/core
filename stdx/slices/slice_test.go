package slices

import (
	"log/slog"
	"sync"
	"testing"

	"github.com/FrogoAI/set"
	"github.com/FrogoAI/testutils"
)

func TestShingle(t *testing.T) {
	testcases := []struct {
		name        string
		text        string
		shingleSize int
		result      set.GenericDataSet[string]
	}{
		{
			name:        "empty",
			text:        "",
			shingleSize: 3,
			result:      set.NewGenericDataSet[string](),
		},
		{
			name:        "zero shingle size",
			text:        "test",
			shingleSize: 0,
			result:      set.NewGenericDataSet[string](),
		},
		{
			name:        "normal strings, 1 size",
			text:        "test",
			shingleSize: 1,
			result: set.NewGenericDataSet[string](
				"t", "e", "s", "t",
			),
		},
		{
			name:        "normal strings, 3 size",
			text:        "test",
			shingleSize: 3,
			result: set.NewGenericDataSet[string](
				"tes", "est",
			),
		},
		{
			name:        "big strings, 3 size",
			text:        "testing",
			shingleSize: 3,
			result: set.NewGenericDataSet[string](
				"tes", "est", "sti", "tin", "ing",
			),
		},
		{
			name:        "short strings, 3 size",
			text:        "t",
			shingleSize: 3,
			result: set.NewGenericDataSet[string](
				"t",
			),
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			res := Shingle(testCase.text, testCase.shingleSize)
			testutils.Equal(t, res, testCase.result)
		})
	}
}

func TestBatchSlice(t *testing.T) {
	size := 102
	maxBatch := 10

	sl := make([]int, 0, maxBatch)
	for i := 0; i < size; i++ {
		sl = append(sl, i)
	}

	var (
		wg sync.WaitGroup
		c  int
	)

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
