package set

import (
	"sort"
	"testing"
)

func sortedStrings(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	sort.Strings(out)
	return out
}

func sortedInts(s []int) []int {
	out := make([]int, len(s))
	copy(out, s)
	sort.Ints(out)
	return out
}

func equalStringSlicesSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa := sortedStrings(a)
	sb := sortedStrings(b)
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func equalIntSlicesSorted(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sa := sortedInts(a)
	sb := sortedInts(b)
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func equalStringSlicesOrdered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestNewGenericDataSet(t *testing.T) {
	cases := []struct {
		name      string
		input     []string
		wantCount int
		wantSlice []string
	}{
		{
			name:      "no arguments",
			input:     nil,
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "empty slice",
			input:     []string{},
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "single element",
			input:     []string{"a"},
			wantCount: 1,
			wantSlice: []string{"a"},
		},
		{
			name:      "multiple distinct elements",
			input:     []string{"a", "b", "c"},
			wantCount: 3,
			wantSlice: []string{"a", "b", "c"},
		},
		{
			name:      "duplicates are deduplicated",
			input:     []string{"a", "a", "b", "b", "c"},
			wantCount: 3,
			wantSlice: []string{"a", "b", "c"},
		},
		{
			name:      "empty string as element",
			input:     []string{""},
			wantCount: 1,
			wantSlice: []string{""},
		},
		{
			name:      "all identical elements",
			input:     []string{"x", "x", "x", "x"},
			wantCount: 1,
			wantSlice: []string{"x"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.input...)
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			if !equalStringSlicesSorted(s.ToSlice(), tc.wantSlice) {
				t.Fatalf("ToSlice: got %v, want %v", s.ToSlice(), tc.wantSlice)
			}
		})
	}
}

func TestNewGenericDataSetInts(t *testing.T) {
	cases := []struct {
		name      string
		input     []int
		wantCount int
	}{
		{
			name:      "int set with zero value",
			input:     []int{0},
			wantCount: 1,
		},
		{
			name:      "int set with negatives",
			input:     []int{-1, -2, -3},
			wantCount: 3,
		},
		{
			name:      "int set with duplicates",
			input:     []int{1, 1, 2, 2},
			wantCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.input...)
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
		})
	}
}

func TestGenericDataSet_Add(t *testing.T) {
	cases := []struct {
		name       string
		initial    []string
		addKeys    []string
		wantCount  int
		wantContain []string
	}{
		{
			name:        "add to empty set",
			initial:     nil,
			addKeys:     []string{"a"},
			wantCount:   1,
			wantContain: []string{"a"},
		},
		{
			name:        "add duplicate key",
			initial:     []string{"a"},
			addKeys:     []string{"a"},
			wantCount:   1,
			wantContain: []string{"a"},
		},
		{
			name:        "add multiple distinct keys",
			initial:     []string{"a"},
			addKeys:     []string{"b", "c"},
			wantCount:   3,
			wantContain: []string{"a", "b", "c"},
		},
		{
			name:        "add empty string",
			initial:     []string{"a"},
			addKeys:     []string{""},
			wantCount:   2,
			wantContain: []string{"a", ""},
		},
		{
			name:        "add same key multiple times",
			initial:     nil,
			addKeys:     []string{"x", "x", "x"},
			wantCount:   1,
			wantContain: []string{"x"},
		},
		{
			name:        "add to non-empty set without overlap",
			initial:     []string{"a", "b"},
			addKeys:     []string{"c", "d"},
			wantCount:   4,
			wantContain: []string{"a", "b", "c", "d"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			for _, k := range tc.wantContain {
				if !s.Contains(k) {
					t.Fatalf("expected set to contain %q", k)
				}
			}
		})
	}
}

func TestGenericDataSet_Delete(t *testing.T) {
	cases := []struct {
		name         string
		initial      []string
		deleteKey    string
		wantCount    int
		shouldAbsent string
	}{
		{
			name:         "delete from empty set",
			initial:      nil,
			deleteKey:    "a",
			wantCount:    0,
			shouldAbsent: "a",
		},
		{
			name:         "delete existing key",
			initial:      []string{"a", "b", "c"},
			deleteKey:    "b",
			wantCount:    2,
			shouldAbsent: "b",
		},
		{
			name:         "delete non-existing key",
			initial:      []string{"a", "b"},
			deleteKey:    "z",
			wantCount:    2,
			shouldAbsent: "z",
		},
		{
			name:         "delete only element",
			initial:      []string{"a"},
			deleteKey:    "a",
			wantCount:    0,
			shouldAbsent: "a",
		},
		{
			name:         "delete empty string key",
			initial:      []string{"", "a"},
			deleteKey:    "",
			wantCount:    1,
			shouldAbsent: "",
		},
		{
			name:         "delete from set with duplicates in constructor",
			initial:      []string{"a", "a", "b"},
			deleteKey:    "a",
			wantCount:    1,
			shouldAbsent: "a",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			s.Delete(tc.deleteKey)
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			if s.Contains(tc.shouldAbsent) {
				t.Fatalf("expected set to not contain %q after delete", tc.shouldAbsent)
			}
		})
	}
}

func TestGenericDataSet_Contains(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		key     string
		want    bool
	}{
		{
			name:    "empty set returns false",
			initial: nil,
			key:     "a",
			want:    false,
		},
		{
			name:    "key present",
			initial: []string{"a", "b"},
			key:     "a",
			want:    true,
		},
		{
			name:    "key absent",
			initial: []string{"a", "b"},
			key:     "c",
			want:    false,
		},
		{
			name:    "empty string key present",
			initial: []string{""},
			key:     "",
			want:    true,
		},
		{
			name:    "empty string key absent",
			initial: []string{"a"},
			key:     "",
			want:    false,
		},
		{
			name:    "single element match",
			initial: []string{"only"},
			key:     "only",
			want:    true,
		},
		{
			name:    "single element no match",
			initial: []string{"only"},
			key:     "other",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			got := s.Contains(tc.key)
			if got != tc.want {
				t.Fatalf("Contains(%q): got %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestGenericDataSet_Count(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    int
	}{
		{
			name:    "empty set",
			initial: nil,
			want:    0,
		},
		{
			name:    "single element",
			initial: []string{"a"},
			want:    1,
		},
		{
			name:    "multiple elements",
			initial: []string{"a", "b", "c"},
			want:    3,
		},
		{
			name:    "duplicates reduce count",
			initial: []string{"a", "a", "b"},
			want:    2,
		},
		{
			name:    "all duplicates",
			initial: []string{"a", "a", "a"},
			want:    1,
		},
		{
			name:    "empty string element",
			initial: []string{""},
			want:    1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			if s.Count() != tc.want {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.want)
			}
		})
	}
}

func TestGenericDataSet_IsEmpty(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    bool
	}{
		{
			name:    "nil input is empty",
			initial: nil,
			want:    true,
		},
		{
			name:    "empty slice is empty",
			initial: []string{},
			want:    true,
		},
		{
			name:    "single element is not empty",
			initial: []string{"a"},
			want:    false,
		},
		{
			name:    "multiple elements is not empty",
			initial: []string{"a", "b"},
			want:    false,
		},
		{
			name:    "empty string element is not empty",
			initial: []string{""},
			want:    false,
		},
		{
			name:    "after add then delete all becomes empty",
			initial: nil,
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			if s.IsEmpty() != tc.want {
				t.Fatalf("IsEmpty: got %v, want %v", s.IsEmpty(), tc.want)
			}
		})
	}
}

func TestGenericDataSet_IsEmpty_AfterMutation(t *testing.T) {
	cases := []struct {
		name      string
		addKeys   []string
		deleteKeys []string
		want      bool
	}{
		{
			name:       "add then delete same key",
			addKeys:    []string{"a"},
			deleteKeys: []string{"a"},
			want:       true,
		},
		{
			name:       "add two delete one",
			addKeys:    []string{"a", "b"},
			deleteKeys: []string{"a"},
			want:       false,
		},
		{
			name:       "add and delete all",
			addKeys:    []string{"a", "b", "c"},
			deleteKeys: []string{"a", "b", "c"},
			want:       true,
		},
		{
			name:       "delete non-existing key",
			addKeys:    []string{"a"},
			deleteKeys: []string{"z"},
			want:       false,
		},
		{
			name:       "no add no delete",
			addKeys:    nil,
			deleteKeys: nil,
			want:       true,
		},
		{
			name:       "add empty string then delete it",
			addKeys:    []string{""},
			deleteKeys: []string{""},
			want:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet[string]()
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			for _, k := range tc.deleteKeys {
				s.Delete(k)
			}
			if s.IsEmpty() != tc.want {
				t.Fatalf("IsEmpty: got %v, want %v", s.IsEmpty(), tc.want)
			}
		})
	}
}

func TestGenericDataSet_ToSlice(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    []string
	}{
		{
			name:    "empty set",
			initial: nil,
			want:    []string{},
		},
		{
			name:    "single element",
			initial: []string{"a"},
			want:    []string{"a"},
		},
		{
			name:    "multiple elements",
			initial: []string{"c", "a", "b"},
			want:    []string{"a", "b", "c"},
		},
		{
			name:    "duplicates in input",
			initial: []string{"a", "a", "b"},
			want:    []string{"a", "b"},
		},
		{
			name:    "empty string element",
			initial: []string{""},
			want:    []string{""},
		},
		{
			name:    "returns correct length",
			initial: []string{"x", "y", "z"},
			want:    []string{"x", "y", "z"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			got := s.ToSlice()
			if !equalStringSlicesSorted(got, tc.want) {
				t.Fatalf("ToSlice: got %v, want %v (order-independent)", sortedStrings(got), sortedStrings(tc.want))
			}
		})
	}
}

func TestGenericDataSet_Union(t *testing.T) {
	cases := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			name: "both empty",
			a:    nil,
			b:    nil,
			want: []string{},
		},
		{
			name: "first empty",
			a:    nil,
			b:    []string{"a", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "second empty",
			a:    []string{"a", "b"},
			b:    nil,
			want: []string{"a", "b"},
		},
		{
			name: "no overlap",
			a:    []string{"a", "b"},
			b:    []string{"c", "d"},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "full overlap",
			a:    []string{"a", "b"},
			b:    []string{"a", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "partial overlap",
			a:    []string{"a", "b", "c"},
			b:    []string{"b", "c", "d"},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "single element sets",
			a:    []string{"a"},
			b:    []string{"b"},
			want: []string{"a", "b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sa := NewGenericDataSet(tc.a...)
			sb := NewGenericDataSet(tc.b...)
			result := sa.Union(sb)
			got := result.ToSlice()
			if !equalStringSlicesSorted(got, tc.want) {
				t.Fatalf("Union: got %v, want %v", sortedStrings(got), sortedStrings(tc.want))
			}
		})
	}
}

func TestGenericDataSet_Union_DoesNotMutateOriginals(t *testing.T) {
	cases := []struct {
		name    string
		a       []string
		b       []string
		aCount  int
		bCount  int
	}{
		{
			name:   "originals unchanged after union",
			a:      []string{"a", "b"},
			b:      []string{"c", "d"},
			aCount: 2,
			bCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sa := NewGenericDataSet(tc.a...)
			sb := NewGenericDataSet(tc.b...)
			_ = sa.Union(sb)
			if sa.Count() != tc.aCount {
				t.Fatalf("original set a mutated: got count %d, want %d", sa.Count(), tc.aCount)
			}
			if sb.Count() != tc.bCount {
				t.Fatalf("original set b mutated: got count %d, want %d", sb.Count(), tc.bCount)
			}
		})
	}
}

func TestGenericDataSet_Intersection(t *testing.T) {
	cases := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{
			name: "both empty",
			a:    nil,
			b:    nil,
			want: []string{},
		},
		{
			name: "first empty",
			a:    nil,
			b:    []string{"a", "b"},
			want: []string{},
		},
		{
			name: "second empty",
			a:    []string{"a", "b"},
			b:    nil,
			want: []string{},
		},
		{
			name: "no overlap",
			a:    []string{"a", "b"},
			b:    []string{"c", "d"},
			want: []string{},
		},
		{
			name: "full overlap",
			a:    []string{"a", "b"},
			b:    []string{"a", "b"},
			want: []string{"a", "b"},
		},
		{
			name: "partial overlap",
			a:    []string{"a", "b", "c"},
			b:    []string{"b", "c", "d"},
			want: []string{"b", "c"},
		},
		{
			name: "single common element",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "d", "e"},
			want: []string{"c"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sa := NewGenericDataSet(tc.a...)
			sb := NewGenericDataSet(tc.b...)
			result := sa.Intersection(sb)
			got := result.ToSlice()
			if !equalStringSlicesSorted(got, tc.want) {
				t.Fatalf("Intersection: got %v, want %v", sortedStrings(got), sortedStrings(tc.want))
			}
		})
	}
}

func TestNewGenericOrderedDataSet(t *testing.T) {
	cases := []struct {
		name      string
		input     []string
		wantCount int
		wantSlice []string
	}{
		{
			name:      "no arguments",
			input:     nil,
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "empty slice",
			input:     []string{},
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "single element",
			input:     []string{"a"},
			wantCount: 1,
			wantSlice: []string{"a"},
		},
		{
			name:      "preserves insertion order",
			input:     []string{"c", "a", "b"},
			wantCount: 3,
			wantSlice: []string{"c", "a", "b"},
		},
		{
			name:      "duplicates are deduplicated preserving first occurrence",
			input:     []string{"a", "b", "a", "c"},
			wantCount: 3,
			wantSlice: []string{"a", "b", "c"},
		},
		{
			name:      "all identical",
			input:     []string{"x", "x", "x"},
			wantCount: 1,
			wantSlice: []string{"x"},
		},
		{
			name:      "empty string element",
			input:     []string{""},
			wantCount: 1,
			wantSlice: []string{""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.input...)
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.wantSlice) {
				t.Fatalf("ToSlice: got %v, want %v", got, tc.wantSlice)
			}
		})
	}
}

func TestGenericOrderedDataSet_Add(t *testing.T) {
	cases := []struct {
		name      string
		initial   []string
		addKeys   []string
		wantCount int
		wantSlice []string
	}{
		{
			name:      "add to empty set",
			initial:   nil,
			addKeys:   []string{"a"},
			wantCount: 1,
			wantSlice: []string{"a"},
		},
		{
			name:      "add duplicate is no-op",
			initial:   []string{"a"},
			addKeys:   []string{"a"},
			wantCount: 1,
			wantSlice: []string{"a"},
		},
		{
			name:      "add preserves order",
			initial:   []string{"b"},
			addKeys:   []string{"a", "c"},
			wantCount: 3,
			wantSlice: []string{"b", "a", "c"},
		},
		{
			name:      "add empty string",
			initial:   []string{"a"},
			addKeys:   []string{""},
			wantCount: 2,
			wantSlice: []string{"a", ""},
		},
		{
			name:      "add same key three times",
			initial:   nil,
			addKeys:   []string{"x", "x", "x"},
			wantCount: 1,
			wantSlice: []string{"x"},
		},
		{
			name:      "add multiple distinct keys",
			initial:   nil,
			addKeys:   []string{"d", "c", "b", "a"},
			wantCount: 4,
			wantSlice: []string{"d", "c", "b", "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.wantSlice) {
				t.Fatalf("ToSlice: got %v, want %v", got, tc.wantSlice)
			}
		})
	}
}

func TestGenericOrderedDataSet_Delete(t *testing.T) {
	cases := []struct {
		name      string
		initial   []string
		deleteKey string
		wantCount int
		wantSlice []string
	}{
		{
			name:      "delete from empty set",
			initial:   nil,
			deleteKey: "a",
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "delete existing element",
			initial:   []string{"a", "b", "c"},
			deleteKey: "b",
			wantCount: 2,
			wantSlice: []string{"a", "c"},
		},
		{
			name:      "delete non-existing element",
			initial:   []string{"a", "b"},
			deleteKey: "z",
			wantCount: 2,
			wantSlice: []string{"a", "b"},
		},
		{
			name:      "delete only element",
			initial:   []string{"a"},
			deleteKey: "a",
			wantCount: 0,
			wantSlice: []string{},
		},
		{
			name:      "delete first element preserves order",
			initial:   []string{"a", "b", "c"},
			deleteKey: "a",
			wantCount: 2,
			wantSlice: []string{"b", "c"},
		},
		{
			name:      "delete last element preserves order",
			initial:   []string{"a", "b", "c"},
			deleteKey: "c",
			wantCount: 2,
			wantSlice: []string{"a", "b"},
		},
		{
			name:      "delete empty string key",
			initial:   []string{"", "a"},
			deleteKey: "",
			wantCount: 1,
			wantSlice: []string{"a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			s.Delete(tc.deleteKey)
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.wantSlice) {
				t.Fatalf("ToSlice: got %v, want %v", got, tc.wantSlice)
			}
		})
	}
}

func TestGenericOrderedDataSet_Last(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    string
	}{
		{
			name:    "single element",
			initial: []string{"a"},
			want:    "a",
		},
		{
			name:    "multiple elements returns last inserted",
			initial: []string{"a", "b", "c"},
			want:    "c",
		},
		{
			name:    "duplicates still return last unique",
			initial: []string{"a", "b", "a"},
			want:    "b",
		},
		{
			name:    "empty string is last",
			initial: []string{"a", ""},
			want:    "",
		},
		{
			name:    "large set returns correct last",
			initial: []string{"1", "2", "3", "4", "5", "6", "7"},
			want:    "7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			got := s.Last()
			if got != tc.want {
				t.Fatalf("Last: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_Last_AfterAdd(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		addKey  string
		want    string
	}{
		{
			name:    "last after adding new element",
			initial: []string{"a", "b"},
			addKey:  "c",
			want:    "c",
		},
		{
			name:    "last after adding duplicate does not change",
			initial: []string{"a", "b"},
			addKey:  "a",
			want:    "b",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			s.Add(tc.addKey)
			got := s.Last()
			if got != tc.want {
				t.Fatalf("Last: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_Contains(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		key     string
		want    bool
	}{
		{
			name:    "empty set returns false",
			initial: nil,
			key:     "a",
			want:    false,
		},
		{
			name:    "key present",
			initial: []string{"a", "b", "c"},
			key:     "b",
			want:    true,
		},
		{
			name:    "key absent",
			initial: []string{"a", "b", "c"},
			key:     "z",
			want:    false,
		},
		{
			name:    "empty string present",
			initial: []string{""},
			key:     "",
			want:    true,
		},
		{
			name:    "empty string absent",
			initial: []string{"a"},
			key:     "",
			want:    false,
		},
		{
			name:    "first element",
			initial: []string{"x", "y", "z"},
			key:     "x",
			want:    true,
		},
		{
			name:    "last element",
			initial: []string{"x", "y", "z"},
			key:     "z",
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			got := s.Contains(tc.key)
			if got != tc.want {
				t.Fatalf("Contains(%q): got %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_Contains_AfterDelete(t *testing.T) {
	cases := []struct {
		name      string
		initial   []string
		deleteKey string
		checkKey  string
		want      bool
	}{
		{
			name:      "deleted key is no longer contained",
			initial:   []string{"a", "b", "c"},
			deleteKey: "b",
			checkKey:  "b",
			want:      false,
		},
		{
			name:      "non-deleted key still contained",
			initial:   []string{"a", "b", "c"},
			deleteKey: "b",
			checkKey:  "a",
			want:      true,
		},
		{
			name:      "delete non-existing key does not affect contains",
			initial:   []string{"a"},
			deleteKey: "z",
			checkKey:  "a",
			want:      true,
		},
		{
			name:      "delete only element then check",
			initial:   []string{"a"},
			deleteKey: "a",
			checkKey:  "a",
			want:      false,
		},
		{
			name:      "delete from empty set then check",
			initial:   nil,
			deleteKey: "a",
			checkKey:  "a",
			want:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			s.Delete(tc.deleteKey)
			got := s.Contains(tc.checkKey)
			if got != tc.want {
				t.Fatalf("Contains(%q) after Delete(%q): got %v, want %v", tc.checkKey, tc.deleteKey, got, tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_ToSlice(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    []string
	}{
		{
			name:    "empty set",
			initial: nil,
			want:    []string{},
		},
		{
			name:    "single element",
			initial: []string{"a"},
			want:    []string{"a"},
		},
		{
			name:    "preserves insertion order",
			initial: []string{"c", "a", "b"},
			want:    []string{"c", "a", "b"},
		},
		{
			name:    "duplicates deduplicated in order",
			initial: []string{"a", "b", "a"},
			want:    []string{"a", "b"},
		},
		{
			name:    "returns a copy not the internal slice",
			initial: []string{"a", "b"},
			want:    []string{"a", "b"},
		},
		{
			name:    "empty string element",
			initial: []string{"x", "", "y"},
			want:    []string{"x", "", "y"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.want) {
				t.Fatalf("ToSlice: got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_ToSlice_ReturnsCopy(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
	}{
		{
			name:    "mutating returned slice does not affect set",
			initial: []string{"a", "b", "c"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			sl := s.ToSlice()
			if len(sl) > 0 {
				sl[0] = "MUTATED"
			}
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.initial) {
				t.Fatalf("internal state changed after mutating ToSlice result: got %v, want %v", got, tc.initial)
			}
		})
	}
}

func TestGenericOrderedDataSet_Count(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		want    int
	}{
		{
			name:    "empty set",
			initial: nil,
			want:    0,
		},
		{
			name:    "single element",
			initial: []string{"a"},
			want:    1,
		},
		{
			name:    "multiple elements",
			initial: []string{"a", "b", "c"},
			want:    3,
		},
		{
			name:    "duplicates reduce count",
			initial: []string{"a", "a", "b"},
			want:    2,
		},
		{
			name:    "all duplicates",
			initial: []string{"x", "x", "x"},
			want:    1,
		},
		{
			name:    "empty string counts",
			initial: []string{"", "a"},
			want:    2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			if s.Count() != tc.want {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.want)
			}
		})
	}
}

func TestGenericOrderedDataSet_Count_AfterMutations(t *testing.T) {
	cases := []struct {
		name       string
		initial    []string
		addKeys    []string
		deleteKeys []string
		want       int
	}{
		{
			name:       "add then delete",
			initial:    nil,
			addKeys:    []string{"a", "b"},
			deleteKeys: []string{"a"},
			want:       1,
		},
		{
			name:       "add duplicates then delete",
			initial:    nil,
			addKeys:    []string{"a", "a", "b"},
			deleteKeys: []string{"a"},
			want:       1,
		},
		{
			name:       "delete non-existing does not change count",
			initial:    []string{"a"},
			addKeys:    nil,
			deleteKeys: []string{"z"},
			want:       1,
		},
		{
			name:       "add after delete",
			initial:    []string{"a"},
			addKeys:    []string{"b"},
			deleteKeys: nil,
			want:       2,
		},
		{
			name:       "delete all elements",
			initial:    []string{"a", "b"},
			addKeys:    nil,
			deleteKeys: []string{"a", "b"},
			want:       0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			for _, k := range tc.deleteKeys {
				s.Delete(k)
			}
			if s.Count() != tc.want {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.want)
			}
		})
	}
}

func TestGenericDataSet_WithIntType(t *testing.T) {
	cases := []struct {
		name        string
		initial     []int
		addKeys     []int
		deleteKeys  []int
		wantCount   int
		wantContain []int
		wantAbsent  []int
	}{
		{
			name:        "int set basic operations",
			initial:     []int{1, 2, 3},
			addKeys:     []int{4},
			deleteKeys:  []int{2},
			wantCount:   3,
			wantContain: []int{1, 3, 4},
			wantAbsent:  []int{2},
		},
		{
			name:        "zero value handling",
			initial:     []int{0},
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   1,
			wantContain: []int{0},
			wantAbsent:  []int{1},
		},
		{
			name:        "negative values",
			initial:     []int{-1, -2},
			addKeys:     []int{-3},
			deleteKeys:  []int{-1},
			wantCount:   2,
			wantContain: []int{-2, -3},
			wantAbsent:  []int{-1},
		},
		{
			name:        "empty int set",
			initial:     nil,
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   0,
			wantContain: nil,
			wantAbsent:  []int{0, 1},
		},
		{
			name:        "duplicate ints",
			initial:     []int{5, 5, 5},
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   1,
			wantContain: []int{5},
			wantAbsent:  []int{6},
		},
		{
			name:        "add and delete same element",
			initial:     nil,
			addKeys:     []int{42},
			deleteKeys:  []int{42},
			wantCount:   0,
			wantContain: nil,
			wantAbsent:  []int{42},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericDataSet(tc.initial...)
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			for _, k := range tc.deleteKeys {
				s.Delete(k)
			}
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			for _, k := range tc.wantContain {
				if !s.Contains(k) {
					t.Fatalf("expected set to contain %d", k)
				}
			}
			for _, k := range tc.wantAbsent {
				if s.Contains(k) {
					t.Fatalf("expected set to not contain %d", k)
				}
			}
		})
	}
}

func TestGenericOrderedDataSet_WithIntType(t *testing.T) {
	cases := []struct {
		name        string
		initial     []int
		addKeys     []int
		deleteKeys  []int
		wantCount   int
		wantSlice   []int
		wantContain []int
		wantAbsent  []int
	}{
		{
			name:        "int ordered set preserves order",
			initial:     []int{3, 1, 2},
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   3,
			wantSlice:   []int{3, 1, 2},
			wantContain: []int{1, 2, 3},
			wantAbsent:  []int{4},
		},
		{
			name:        "zero value in ordered set",
			initial:     []int{0, 1},
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   2,
			wantSlice:   []int{0, 1},
			wantContain: []int{0, 1},
			wantAbsent:  []int{2},
		},
		{
			name:        "delete preserves order",
			initial:     []int{10, 20, 30},
			addKeys:     nil,
			deleteKeys:  []int{20},
			wantCount:   2,
			wantSlice:   []int{10, 30},
			wantContain: []int{10, 30},
			wantAbsent:  []int{20},
		},
		{
			name:        "add preserves order after initial",
			initial:     []int{1},
			addKeys:     []int{3, 2},
			deleteKeys:  nil,
			wantCount:   3,
			wantSlice:   []int{1, 3, 2},
			wantContain: []int{1, 2, 3},
			wantAbsent:  []int{4},
		},
		{
			name:        "negative ints ordered",
			initial:     []int{-3, -1, -2},
			addKeys:     nil,
			deleteKeys:  nil,
			wantCount:   3,
			wantSlice:   []int{-3, -1, -2},
			wantContain: []int{-3, -1, -2},
			wantAbsent:  []int{0},
		},
		{
			name:        "re-add after delete appends to end",
			initial:     []int{1, 2, 3},
			addKeys:     nil,
			deleteKeys:  []int{2},
			wantCount:   2,
			wantSlice:   []int{1, 3},
			wantContain: []int{1, 3},
			wantAbsent:  []int{2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			for _, k := range tc.addKeys {
				s.Add(k)
			}
			for _, k := range tc.deleteKeys {
				s.Delete(k)
			}
			if s.Count() != tc.wantCount {
				t.Fatalf("Count: got %d, want %d", s.Count(), tc.wantCount)
			}
			got := s.ToSlice()
			if len(got) != len(tc.wantSlice) {
				t.Fatalf("ToSlice length: got %d, want %d", len(got), len(tc.wantSlice))
			}
			for i := range got {
				if got[i] != tc.wantSlice[i] {
					t.Fatalf("ToSlice[%d]: got %d, want %d", i, got[i], tc.wantSlice[i])
				}
			}
			for _, k := range tc.wantContain {
				if !s.Contains(k) {
					t.Fatalf("expected set to contain %d", k)
				}
			}
			for _, k := range tc.wantAbsent {
				if s.Contains(k) {
					t.Fatalf("expected set to not contain %d", k)
				}
			}
		})
	}
}

func TestGenericOrderedDataSet_AddAfterDelete(t *testing.T) {
	cases := []struct {
		name      string
		initial   []string
		deleteKey string
		addKey    string
		wantSlice []string
	}{
		{
			name:      "re-add deleted element goes to end",
			initial:   []string{"a", "b", "c"},
			deleteKey: "b",
			addKey:    "b",
			wantSlice: []string{"a", "c", "b"},
		},
		{
			name:      "add new element after delete",
			initial:   []string{"a", "b"},
			deleteKey: "a",
			addKey:    "c",
			wantSlice: []string{"b", "c"},
		},
		{
			name:      "re-add only element",
			initial:   []string{"a"},
			deleteKey: "a",
			addKey:    "a",
			wantSlice: []string{"a"},
		},
		{
			name:      "delete first then re-add",
			initial:   []string{"x", "y", "z"},
			deleteKey: "x",
			addKey:    "x",
			wantSlice: []string{"y", "z", "x"},
		},
		{
			name:      "delete last then re-add",
			initial:   []string{"x", "y", "z"},
			deleteKey: "z",
			addKey:    "z",
			wantSlice: []string{"x", "y", "z"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewGenericOrderedDataSet(tc.initial...)
			s.Delete(tc.deleteKey)
			s.Add(tc.addKey)
			got := s.ToSlice()
			if !equalStringSlicesOrdered(got, tc.wantSlice) {
				t.Fatalf("ToSlice after delete+add: got %v, want %v", got, tc.wantSlice)
			}
		})
	}
}
