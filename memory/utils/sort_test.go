package utils

import (
	"math/rand"
	"testing"

	"github.com/InsideGallery/core/memory/comparator"
)

func TestSort(t *testing.T) {
	type testCase struct {
		name       string
		values     []interface{}
		comparator comparator.Comparator
		want       []interface{}
	}

	cases := []testCase{
		{
			name:       "empty slice",
			values:     []interface{}{},
			comparator: comparator.IntComparator,
			want:       []interface{}{},
		},
		{
			name:       "single element",
			values:     []interface{}{42},
			comparator: comparator.IntComparator,
			want:       []interface{}{42},
		},
		{
			name:       "already sorted ints",
			values:     []interface{}{1, 2, 3, 4, 5},
			comparator: comparator.IntComparator,
			want:       []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:       "reverse sorted ints",
			values:     []interface{}{5, 4, 3, 2, 1},
			comparator: comparator.IntComparator,
			want:       []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:       "duplicate ints",
			values:     []interface{}{3, 1, 2, 1, 3, 2},
			comparator: comparator.IntComparator,
			want:       []interface{}{1, 1, 2, 2, 3, 3},
		},
		{
			name:       "all same ints",
			values:     []interface{}{7, 7, 7, 7},
			comparator: comparator.IntComparator,
			want:       []interface{}{7, 7, 7, 7},
		},
		{
			name:       "negative ints",
			values:     []interface{}{-3, 0, -1, 5, -2},
			comparator: comparator.IntComparator,
			want:       []interface{}{-3, -2, -1, 0, 5},
		},
		{
			name:       "two elements swapped",
			values:     []interface{}{2, 1},
			comparator: comparator.IntComparator,
			want:       []interface{}{1, 2},
		},
		{
			name:       "strings unsorted",
			values:     []interface{}{"delta", "alpha", "charlie", "bravo"},
			comparator: comparator.StringComparator,
			want:       []interface{}{"alpha", "bravo", "charlie", "delta"},
		},
		{
			name:       "strings with empty string",
			values:     []interface{}{"b", "", "a"},
			comparator: comparator.StringComparator,
			want:       []interface{}{"", "a", "b"},
		},
		{
			name:       "strings all same",
			values:     []interface{}{"x", "x", "x"},
			comparator: comparator.StringComparator,
			want:       []interface{}{"x", "x", "x"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Sort(tc.values, tc.comparator)

			if len(tc.values) != len(tc.want) {
				t.Fatalf("length mismatch: got %d, want %d", len(tc.values), len(tc.want))
			}

			for i := range tc.values {
				if tc.values[i] != tc.want[i] {
					t.Errorf("index %d: got %v, want %v", i, tc.values[i], tc.want[i])
				}
			}
		})
	}
}

func TestSort_CustomComparator(t *testing.T) {
	type user struct {
		id   int
		name string
	}

	byID := func(a, b interface{}) int {
		c1 := a.(user)

		c2 := b.(user)
		switch {
		case c1.id > c2.id:
			return 1
		case c1.id < c2.id:
			return -1
		default:
			return 0
		}
	}

	cases := []struct {
		name   string
		values []interface{}
		want   []int
	}{
		{
			name:   "sort structs by id",
			values: []interface{}{user{4, "d"}, user{1, "a"}, user{3, "c"}, user{2, "b"}},
			want:   []int{1, 2, 3, 4},
		},
		{
			name:   "single struct",
			values: []interface{}{user{1, "a"}},
			want:   []int{1},
		},
		{
			name:   "empty structs",
			values: []interface{}{},
			want:   []int{},
		},
		{
			name:   "already sorted structs",
			values: []interface{}{user{1, "a"}, user{2, "b"}, user{3, "c"}},
			want:   []int{1, 2, 3},
		},
		{
			name:   "reverse sorted structs",
			values: []interface{}{user{3, "c"}, user{2, "b"}, user{1, "a"}},
			want:   []int{1, 2, 3},
		},
		{
			name:   "duplicate ids",
			values: []interface{}{user{2, "x"}, user{1, "y"}, user{2, "z"}},
			want:   []int{1, 2, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Sort(tc.values, byID)

			if len(tc.values) != len(tc.want) {
				t.Fatalf("length mismatch: got %d, want %d", len(tc.values), len(tc.want))
			}

			for i := range tc.values {
				if tc.values[i].(user).id != tc.want[i] {
					t.Errorf("index %d: got id %d, want %d", i, tc.values[i].(user).id, tc.want[i])
				}
			}
		})
	}
}

func TestSort_LargeRandom(t *testing.T) {
	cases := []struct {
		name string
		size int
	}{
		{name: "100 random ints", size: 100},
		{name: "1000 random ints", size: 1000},
		{name: "10000 random ints", size: 10000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ints := make([]interface{}, tc.size)
			for i := range ints {
				ints[i] = rand.Int() //nolint:gosec
			}

			Sort(ints, comparator.IntComparator)

			for i := 1; i < len(ints); i++ {
				if ints[i-1].(int) > ints[i].(int) {
					t.Fatalf("not sorted at index %d: %d > %d", i, ints[i-1].(int), ints[i].(int))
				}
			}
		})
	}
}

func TestSortable_Len(t *testing.T) {
	cases := []struct {
		name   string
		values []interface{}
		want   int
	}{
		{name: "empty", values: []interface{}{}, want: 0},
		{name: "nil", values: nil, want: 0},
		{name: "one", values: []interface{}{1}, want: 1},
		{name: "five", values: []interface{}{1, 2, 3, 4, 5}, want: 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := sortable{values: tc.values, comparator: comparator.IntComparator}
			if got := s.Len(); got != tc.want {
				t.Errorf("Len() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSortable_Swap(t *testing.T) {
	cases := []struct {
		name   string
		values []interface{}
		i, j   int
		wantI  interface{}
		wantJ  interface{}
	}{
		{name: "swap first and last", values: []interface{}{1, 2, 3}, i: 0, j: 2, wantI: 3, wantJ: 1},
		{name: "swap same index", values: []interface{}{1, 2, 3}, i: 1, j: 1, wantI: 2, wantJ: 2},
		{name: "swap adjacent", values: []interface{}{10, 20}, i: 0, j: 1, wantI: 20, wantJ: 10},
		{name: "swap with strings", values: []interface{}{"a", "b", "c"}, i: 0, j: 2, wantI: "c", wantJ: "a"},
		{name: "swap middle elements", values: []interface{}{1, 2, 3, 4, 5}, i: 1, j: 3, wantI: 4, wantJ: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := sortable{values: tc.values, comparator: comparator.IntComparator}
			s.Swap(tc.i, tc.j)

			if s.values[tc.i] != tc.wantI {
				t.Errorf("after Swap values[%d] = %v, want %v", tc.i, s.values[tc.i], tc.wantI)
			}

			if s.values[tc.j] != tc.wantJ {
				t.Errorf("after Swap values[%d] = %v, want %v", tc.j, s.values[tc.j], tc.wantJ)
			}
		})
	}
}

func TestSortable_Less(t *testing.T) {
	cases := []struct {
		name   string
		values []interface{}
		i, j   int
		want   bool
	}{
		{name: "first less than second", values: []interface{}{1, 2}, i: 0, j: 1, want: true},
		{name: "first greater than second", values: []interface{}{2, 1}, i: 0, j: 1, want: false},
		{name: "equal elements", values: []interface{}{3, 3}, i: 0, j: 1, want: false},
		{name: "negative less than positive", values: []interface{}{-1, 1}, i: 0, j: 1, want: true},
		{name: "zero and positive", values: []interface{}{0, 1}, i: 0, j: 1, want: true},
		{name: "zero and negative", values: []interface{}{0, -1}, i: 0, j: 1, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := sortable{values: tc.values, comparator: comparator.IntComparator}
			if got := s.Less(tc.i, tc.j); got != tc.want {
				t.Errorf("Less(%d, %d) = %v, want %v", tc.i, tc.j, got, tc.want)
			}
		})
	}
}

func BenchmarkSort_Random(b *testing.B) {
	b.StopTimer()

	ints := make([]interface{}, 100000)
	for i := range ints {
		ints[i] = rand.Int() //nolint:gosec
	}

	b.StartTimer()
	Sort(ints, comparator.IntComparator)
	b.StopTimer()
}
