package sortedset

import (
	"errors"
	"reflect"
	"testing"

	"github.com/InsideGallery/core/memory/comparator"
)

func TestSortedSetAdditionalOperations(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "peek pop remove and contains",
			run:  assertSortedSetPeekRemoveContains,
		},
		{
			name: "rank ranges support forward reverse and remove",
			run:  assertSortedSetRankRanges,
		},
		{
			name: "key ranges support limits exclusions reverse and remove",
			run:  assertSortedSetKeyRanges,
		},
		{
			name: "dump and restore errors are returned",
			run:  assertSortedSetDumpRestoreErrors,
		},
		{
			name: "lookup helpers handle existing missing and empty values",
			run:  assertSortedSetLookupHelpers,
		},
		{
			name: "dump and restore round trip",
			run:  assertSortedSetDumpRestoreRoundTrip,
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func assertSortedSetPeekRemoveContains(t *testing.T) {
	t.Helper()

	set := newIntSet()
	upsertIntValues(set, map[string]int{"b": 2, "a": 1, "c": 3})

	if !set.Contains("b") {
		t.Fatal("set should contain b")
	}

	if added := set.Upsert(2, "b"); added {
		t.Fatal("updating existing value should return false")
	}

	if added := set.Upsert(4, "b"); added {
		t.Fatal("moving existing value should return false")
	}

	if minNode := set.PeekMin(); minNode.Value() != "a" {
		t.Fatalf("min = %q, want a", minNode.Value())
	}

	if maxNode := set.PeekMax(); maxNode.Value() != "b" {
		t.Fatalf("max = %q, want b", maxNode.Value())
	}

	if removed := set.Remove("missing"); removed != nil {
		t.Fatal("missing value should not be removed")
	}

	if removed := set.Remove("a"); removed.Value() != "a" {
		t.Fatalf("removed = %q, want a", removed.Value())
	}

	if removed := set.Remove("b"); removed.Value() != "b" {
		t.Fatalf("removed = %q, want b", removed.Value())
	}

	if set.GetCount() != 1 {
		t.Fatalf("count = %d, want 1", set.GetCount())
	}
}

func assertSortedSetRankRanges(t *testing.T) {
	t.Helper()

	set := rankedIntSet()

	if got := nodeValues(set.GetByRankRange(2, 3, false)); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("rank range = %v", got)
	}

	if got := nodeValues(set.GetByRankRange(4, 2, false)); !reflect.DeepEqual(got, []string{"d", "c", "b"}) {
		t.Fatalf("reverse rank range = %v", got)
	}

	if got := set.GetByRank(-1, false).Value(); got != "d" {
		t.Fatalf("rank -1 = %q, want d", got)
	}

	if got := nodeValues(set.GetTop(2, false)); !reflect.DeepEqual(got, []string{"d", "c"}) {
		t.Fatalf("top = %v", got)
	}

	if got := nodeValues(set.GetRTop(2, false)); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("reverse top = %v", got)
	}

	removed := set.GetByRankRange(1, 2, true)
	if got := nodeValues(removed); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("removed = %v", got)
	}

	if set.GetCount() != 2 {
		t.Fatalf("count = %d, want 2", set.GetCount())
	}
}

func assertSortedSetKeyRanges(t *testing.T) {
	t.Helper()

	set := rankedIntSet()

	if got := nodeValues(set.GetByKeyRange(1, 3, nil)); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("key range = %v", got)
	}

	exclusive := set.GetByKeyRange(1, 3, &GetByKeyRangeOptions{
		ExcludeStart: true,
		ExcludeEnd:   true,
	})
	if got := nodeValues(exclusive); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("exclusive = %v", got)
	}

	limited := set.GetByKeyRange(1, 4, &GetByKeyRangeOptions{Limit: 2})
	if got := nodeValues(limited); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("limited = %v", got)
	}

	reversed := set.GetByKeyRange(4, 2, nil)
	if got := nodeValues(reversed); !reflect.DeepEqual(got, []string{"d", "c", "b"}) {
		t.Fatalf("reversed = %v", got)
	}

	removed := set.GetByKeyRange(1, 2, &GetByKeyRangeOptions{Remove: true})
	if got := nodeValues(removed); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("removed = %v", got)
	}

	if set.Contains("a") || set.Contains("b") {
		t.Fatal("removed values should not remain")
	}
}

func assertSortedSetDumpRestoreErrors(t *testing.T) {
	t.Helper()

	expectedErr := errors.New("dump failed")
	set := rankedIntSet()

	_, err := set.Dump(func(int, string) (string, string, error) {
		return "", "", expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("dump err = %v, want %v", err, expectedErr)
	}

	if err := set.Restore(nil, "{"); err == nil {
		t.Fatal("expected invalid json error")
	}

	err = set.Restore(func(string, []string) (int, []string, error) {
		return 0, nil, expectedErr
	}, `{"1":["a"]}`)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("restore err = %v, want %v", err, expectedErr)
	}
}

func assertSortedSetLookupHelpers(t *testing.T) {
	t.Helper()

	empty := newIntSet()
	if empty.PeekMin() != nil || empty.PeekMax() != nil {
		t.Fatal("empty set should not have min or max")
	}

	if empty.GetByRank(1, false) != nil {
		t.Fatal("empty set should not return rank")
	}

	set := rankedIntSet()
	if got := set.GetByValue("c").Key(); got != 3 {
		t.Fatalf("key = %d, want 3", got)
	}

	_ = set.FindRank("c")

	if got := set.FindRank("missing"); got != 0 {
		t.Fatalf("missing rank = %d, want 0", got)
	}
}

func assertSortedSetDumpRestoreRoundTrip(t *testing.T) {
	t.Helper()

	set := rankedIntSet()
	dump, err := set.Dump(func(key int, value string) (string, string, error) {
		return string(rune('0' + key)), value, nil
	})
	if err != nil {
		t.Fatalf("dump: %v", err)
	}

	restored := newIntSet()
	err = restored.Restore(func(key string, values []string) (int, []string, error) {
		return int([]rune(key)[0] - '0'), values, nil
	}, dump)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	if got := nodeValues(restored.GetByRankRange(1, 4, false)); !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("restored = %v", got)
	}
}

func newIntSet() *SortedSet[int, string] {
	return NewSortedSet[int, string](comparator.IntComparator)
}

func rankedIntSet() *SortedSet[int, string] {
	set := newIntSet()
	upsertIntValues(set, map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})

	return set
}

func upsertIntValues(set *SortedSet[int, string], values map[string]int) {
	for value, key := range values {
		set.Upsert(key, value)
	}
}

func nodeValues(nodes []*Node[int, string]) []string {
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, node.Value())
	}

	return values
}
