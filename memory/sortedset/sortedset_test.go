package sortedset

import (
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/InsideGallery/core/memory/comparator"
	"github.com/InsideGallery/core/testutils"
)

func TestTimeQueue(t *testing.T) {
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
