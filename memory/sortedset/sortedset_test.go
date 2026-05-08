package sortedset

import (
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/InsideGallery/core/memory/comparator"
	"github.com/InsideGallery/core/testutils"
)

const (
	sortedSetLargeStart  = -512
	sortedSetLargeEnd    = 512
	sortedSetOnlyValue   = "only"
	sortedSetReaderCount = 8
)

func TestTimeQueue(t *testing.T) {
	t.Parallel()

	set := NewSortedSet[time.Time, any](comparator.TimeComparator)
	for i := -1000; i <= 10000; i++ {
		set.Upsert(time.Now().Add(time.Duration(i)*time.Minute), i)
	}

	set.Upsert(time.Now().Add(time.Minute), "test4")
	set.Upsert(time.Now().Add(time.Second), "test3")
	set.Upsert(time.Now().Add(-time.Minute), "test1")
	set.Upsert(time.Now().Add(-time.Hour), "test0")
	set.Upsert(time.Now(), "test2")

	set.GetUntilKey(time.Now().Add(-5*time.Minute), true)
	values := set.GetUntilKey(time.Now(), true)
	expectedValues := []interface{}{
		-4, -3, -2, -1, "test1", 0, "test2",
	}

	testutils.Equal(t, values, expectedValues)
}

func TestDump(t *testing.T) {
	t.Parallel()

	set := NewSortedSet[time.Time, any](comparator.TimeComparator)
	set.Upsert(time.Now().Add(time.Minute), "test4")
	set.Upsert(time.Now().Add(time.Second), "test3")
	set.Upsert(time.Now().Add(-time.Minute), "test1")
	set.Upsert(time.Now().Add(-time.Hour), "test0")
	set.Upsert(time.Now(), "test2")

	dump, err := set.Dump(func(key time.Time, values any) (string, string, error) {
		return key.Format(time.RFC3339), cast.ToString(values), nil
	})
	testutils.Equal(t, err, nil)

	set = NewSortedSet[time.Time, any](comparator.TimeComparator)
	err = set.Restore(func(key string, values []string) (time.Time, []any, error) {
		rValues := make([]any, len(values))
		for i, v := range values {
			rValues[i] = v
		}

		rKey, err := time.Parse(time.RFC3339, key)

		return rKey, rValues, err
	}, dump)
	testutils.Equal(t, err, nil)

	set.GetUntilKey(time.Now().Add(-5*time.Minute), true)
	values := set.GetUntilKey(time.Now(), true)
	expectedValues := []interface{}{
		"test1", "test2",
	}

	testutils.Equal(t, values, expectedValues)
}

func TestSortedSetBoundaryConditions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "empty set returns no nodes",
			run: func(t *testing.T) {
				t.Helper()

				set := NewSortedSet[int, string](comparator.IntComparator)

				if set.GetCount() != 0 {
					t.Fatalf("count = %d, want 0", set.GetCount())
				}

				if set.PeekMin() != nil {
					t.Fatal("min should be nil")
				}

				if set.PeekMax() != nil {
					t.Fatal("max should be nil")
				}

				if set.PopMin() != nil {
					t.Fatal("pop min should be nil")
				}

				if set.PopMax() != nil {
					t.Fatal("pop max should be nil")
				}

				if set.GetByRank(1, false) != nil {
					t.Fatal("rank should be nil")
				}

				if nodes := set.GetByKeyRange(0, 10, nil); len(nodes) != 0 {
					t.Fatalf("range len = %d, want 0", len(nodes))
				}
			},
		},
		{
			name: "single element is min max and rank",
			run: func(t *testing.T) {
				t.Helper()

				set := NewSortedSet[int, string](comparator.IntComparator)
				if added := set.Upsert(7, sortedSetOnlyValue); !added {
					t.Fatal("first upsert should add")
				}

				if set.GetCount() != 1 {
					t.Fatalf("count = %d, want 1", set.GetCount())
				}

				if got := set.PeekMin().Value(); got != sortedSetOnlyValue {
					t.Fatalf("min = %q, want %s", got, sortedSetOnlyValue)
				}

				if got := set.PeekMax().Value(); got != sortedSetOnlyValue {
					t.Fatalf("max = %q, want %s", got, sortedSetOnlyValue)
				}

				if got := set.GetByRank(1, false).Value(); got != sortedSetOnlyValue {
					t.Fatalf("rank = %q, want %s", got, sortedSetOnlyValue)
				}

				if removed := set.Remove(sortedSetOnlyValue); removed == nil || removed.Value() != sortedSetOnlyValue {
					t.Fatalf("removed = %v, want %s", removed, sortedSetOnlyValue)
				}

				if set.GetCount() != 0 {
					t.Fatalf("count after remove = %d, want 0", set.GetCount())
				}
			},
		},
		{
			name: "duplicate keys keep distinct values",
			run: func(t *testing.T) {
				t.Helper()

				set := NewSortedSet[int, string](comparator.IntComparator)
				values := []string{"first", "second", "third"}

				for _, value := range values {
					if added := set.Upsert(10, value); !added {
						t.Fatalf("value %q should be added", value)
					}
				}

				if set.GetCount() != len(values) {
					t.Fatalf("count = %d, want %d", set.GetCount(), len(values))
				}

				got := sortedSetBoundaryStringValues(set.GetByKeyRange(10, 10, nil))
				if !slices.Equal(got, values) {
					t.Fatalf("values = %v, want %v", got, values)
				}
			},
		},
		{
			name: "large key range keeps sorted boundaries",
			run: func(t *testing.T) {
				t.Helper()

				set := NewSortedSet[int, int](comparator.IntComparator)
				for key := sortedSetLargeStart; key <= sortedSetLargeEnd; key++ {
					set.Upsert(key, key)
				}

				wantCount := sortedSetLargeEnd - sortedSetLargeStart + 1
				if set.GetCount() != wantCount {
					t.Fatalf("count = %d, want %d", set.GetCount(), wantCount)
				}

				nodes := set.GetByKeyRange(sortedSetLargeStart, sortedSetLargeEnd, nil)
				if len(nodes) != wantCount {
					t.Fatalf("range len = %d, want %d", len(nodes), wantCount)
				}

				if got := nodes[0].Key(); got != sortedSetLargeStart {
					t.Fatalf("first key = %d, want %d", got, sortedSetLargeStart)
				}

				if got := nodes[len(nodes)-1].Key(); got != sortedSetLargeEnd {
					t.Fatalf("last key = %d, want %d", got, sortedSetLargeEnd)
				}
			},
		},
	}

	for _, test := range cases {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.run(t)
		})
	}
}

func TestSortedSetGetByKeyRangeBoundaryCases(t *testing.T) {
	t.Parallel()

	type entry struct {
		key   int
		value string
	}

	cases := []struct {
		name    string
		entries []entry
		start   int
		end     int
		options *GetByKeyRangeOptions
		want    []string
	}{
		{
			name: "empty range excludes equal endpoints",
			entries: []entry{
				{key: 1, value: "a"},
				{key: 2, value: "b"},
				{key: 3, value: "c"},
			},
			start: 2,
			end:   2,
			options: &GetByKeyRangeOptions{
				ExcludeStart: true,
				ExcludeEnd:   true,
			},
			want: []string{},
		},
		{
			name: "range below min returns empty",
			entries: []entry{
				{key: 10, value: "a"},
				{key: 20, value: "b"},
			},
			start: 1,
			end:   9,
			want:  []string{},
		},
		{
			name: "range above max returns empty",
			entries: []entry{
				{key: 10, value: "a"},
				{key: 20, value: "b"},
			},
			start: 21,
			end:   30,
			want:  []string{},
		},
		{
			name: "range crossing both ends returns all values",
			entries: []entry{
				{key: 10, value: "a"},
				{key: 20, value: "b"},
				{key: 30, value: "c"},
			},
			start: 1,
			end:   40,
			want:  []string{"a", "b", "c"},
		},
		{
			name: "duplicate keys are all returned",
			entries: []entry{
				{key: 10, value: "a"},
				{key: 20, value: "b"},
				{key: 20, value: "b2"},
				{key: 20, value: "b3"},
				{key: 30, value: "c"},
			},
			start: 20,
			end:   20,
			want:  []string{"b", "b2", "b3"},
		},
	}

	for _, test := range cases {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			set := NewSortedSet[int, string](comparator.IntComparator)
			for _, item := range test.entries {
				set.Upsert(item.key, item.value)
			}

			got := sortedSetBoundaryStringValues(set.GetByKeyRange(test.start, test.end, test.options))
			if !slices.Equal(got, test.want) {
				t.Fatalf("values = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSortedSetConcurrentReads(t *testing.T) {
	t.Parallel()

	set := NewSortedSet[int, int](comparator.IntComparator)
	for key := 0; key <= sortedSetLargeEnd; key++ {
		set.Upsert(key, key)
	}

	errCh := make(chan error, sortedSetReaderCount)

	var wg sync.WaitGroup
	wg.Add(sortedSetReaderCount)

	for range sortedSetReaderCount {
		go func() {
			defer wg.Done()

			for value := 0; value <= sortedSetLargeEnd; value++ {
				node := set.GetByValue(value)
				if node == nil {
					errCh <- fmt.Errorf("value %d missing", value)
					return
				}

				if node.Key() != value {
					errCh <- fmt.Errorf("key = %d, want %d", node.Key(), value)
					return
				}
			}

			if minNode := set.PeekMin(); minNode == nil || minNode.Key() != 0 {
				errCh <- fmt.Errorf("min = %v, want key 0", minNode)
				return
			}

			if maxNode := set.PeekMax(); maxNode == nil || maxNode.Key() != sortedSetLargeEnd {
				errCh <- fmt.Errorf("max = %v, want key %d", maxNode, sortedSetLargeEnd)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func sortedSetBoundaryStringValues(nodes []*Node[int, string]) []string {
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, node.Value())
	}

	return values
}

/*
BenchmarkUpsert-12               1000000              1065 ns/op             241 B/op          4 allocs/op
BenchmarkUpsertAlt-12            3403590               351.5 ns/op           335 B/op          3 allocs/op
BenchmarkGetUntilKey-12            44418             26979 ns/op           32784 B/op         13 allocs/op
BenchmarkGetUntilKeyAlt-12         70933             17045 ns/op           32832 B/op         13 allocs/op
BenchmarkTop-12                  4050680               314.8 ns/op           120 B/op          4 allocs/op
BenchmarkTopWithRemove-12       32379015                34.64 ns/op            0 B/op          0 allocs/op
*/

var resultList []*Node[time.Time, any]

func BenchmarkUpsert(b *testing.B) {
	s := NewSortedSet[time.Time, any](comparator.TimeComparator)

	for i := 0; i < b.N; i++ {
		j := i
		if j%2 == 0 {
			j *= -1
		}

		s.Upsert(time.Now().Add(time.Duration(j)*time.Minute), i)
	}
}

func BenchmarkGetUntilKey(b *testing.B) {
	b.StopTimer()

	s := NewSortedSet[time.Time, any](comparator.TimeComparator)
	for i := -1000; i <= 10000; i++ {
		s.Upsert(time.Now().Add(time.Duration(i)*time.Minute), i)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = s.GetUntilKey(time.Now(), false)
	}
}

func BenchmarkTop(b *testing.B) {
	b.StopTimer()

	s := NewSortedSet[time.Time, any](comparator.TimeComparator)
	for i := -1000; i <= 10000; i++ {
		s.Upsert(time.Now().Add(time.Duration(i)*time.Minute), i)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		resultList = s.GetTop(5, false)
	}
}

func BenchmarkTopWithRemove(b *testing.B) {
	b.StopTimer()

	s := NewSortedSet[time.Time, any](comparator.TimeComparator)
	for i := -1000; i <= 10000; i++ {
		s.Upsert(time.Now().Add(time.Duration(i)*time.Minute), i)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		resultList = s.GetTop(5, true)
	}
}
