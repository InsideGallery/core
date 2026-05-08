package orderedmap

import (
	"sort"
	"sync"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestOrderedmap(t *testing.T) {
	o := &OrderedMap[string, string]{}
	o.Add("A", "1")
	o.Add("B", "2")
	o.Add("C", "3")
	o.Add("E", "4")
	o.Remove("C")
	testutils.Equal(t, o.Get("A"), "1")
	testutils.Equal(t, o.Get("C"), "")
	keys, values := o.GetAll()
	testutils.Equal(t, len(keys), 3)
	testutils.Equal(t, len(values), 3)
	testutils.Equal(t, keys, []string{"A", "B", "E"})
	testutils.Equal(t, values, []string{"1", "2", "4"})
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		addKey     string
		addVal     int
		wantSize   int
		wantKeys   []string
		wantValues []int
	}{
		{
			name:       "add to empty map",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			addKey:     "a",
			addVal:     1,
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{1},
		},
		{
			name: "add new key to existing map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			addKey:     "b",
			addVal:     2,
			wantSize:   2,
			wantKeys:   []string{"a", "b"},
			wantValues: []int{1, 2},
		},
		{
			name: "overwrite existing key preserves order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			addKey:     "a",
			addVal:     99,
			wantSize:   2,
			wantKeys:   []string{"a", "b"},
			wantValues: []int{99, 2},
		},
		{
			name: "overwrite middle key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("x", 10)
				o.Add("y", 20)
				o.Add("z", 30)

				return o
			},
			addKey:     "y",
			addVal:     200,
			wantSize:   3,
			wantKeys:   []string{"x", "y", "z"},
			wantValues: []int{10, 200, 30},
		},
		{
			name:       "add with zero value key",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			addKey:     "",
			addVal:     42,
			wantSize:   1,
			wantKeys:   []string{""},
			wantValues: []int{42},
		},
		{
			name:       "add with zero value val",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			addKey:     "k",
			addVal:     0,
			wantSize:   1,
			wantKeys:   []string{"k"},
			wantValues: []int{0},
		},
		{
			name: "duplicate add does not duplicate key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("a", 1)
				o.Add("a", 1)

				return o
			},
			addKey:     "a",
			addVal:     1,
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{1},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			o.Add(tc.addKey, tc.addVal)
			testutils.Equal(t, o.Size(), tc.wantSize)
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *OrderedMap[string, int]
		key   string
		want  int
	}{
		{
			name:  "get from empty map returns zero value",
			setup: func() *OrderedMap[string, int] { return &OrderedMap[string, int]{values: map[string]int{}} },
			key:   "missing",
			want:  0,
		},
		{
			name: "get existing key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 42)

				return o
			},
			key:  "a",
			want: 42,
		},
		{
			name: "get non-existent key from non-empty map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			key:  "b",
			want: 0,
		},
		{
			name: "get after overwrite returns new value",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("a", 99)

				return o
			},
			key:  "a",
			want: 99,
		},
		{
			name: "get after remove returns zero value",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Remove("a")

				return o
			},
			key:  "a",
			want: 0,
		},
		{
			name: "get empty string key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("", 7)

				return o
			},
			key:  "",
			want: 7,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			testutils.Equal(t, o.Get(tc.key), tc.want)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		removeKey  string
		wantSize   int
		wantKeys   []string
		wantValues []int
	}{
		{
			name: "remove from empty map does not panic",
			setup: func() *OrderedMap[string, int] {
				return &OrderedMap[string, int]{values: map[string]int{}, keys: []string{}}
			},
			removeKey:  "x",
			wantSize:   0,
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "remove non-existent key from non-empty map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			removeKey:  "z",
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{1},
		},
		{
			name: "remove only element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			removeKey:  "a",
			wantSize:   0,
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "remove first element preserves order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			removeKey:  "a",
			wantSize:   2,
			wantKeys:   []string{"b", "c"},
			wantValues: []int{2, 3},
		},
		{
			name: "remove last element preserves order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			removeKey:  "c",
			wantSize:   2,
			wantKeys:   []string{"a", "b"},
			wantValues: []int{1, 2},
		},
		{
			name: "remove middle element preserves order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			removeKey:  "b",
			wantSize:   2,
			wantKeys:   []string{"a", "c"},
			wantValues: []int{1, 3},
		},
		{
			name: "remove same key twice",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Remove("a")

				return o
			},
			removeKey:  "a",
			wantSize:   1,
			wantKeys:   []string{"b"},
			wantValues: []int{2},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			o.Remove(tc.removeKey)
			testutils.Equal(t, o.Size(), tc.wantSize)
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *OrderedMap[string, int]
		key   string
		want  bool
	}{
		{
			name:  "exists on empty map",
			setup: func() *OrderedMap[string, int] { return &OrderedMap[string, int]{values: map[string]int{}} },
			key:   "a",
			want:  false,
		},
		{
			name: "exists for present key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			key:  "a",
			want: true,
		},
		{
			name: "exists for absent key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			key:  "b",
			want: false,
		},
		{
			name: "exists after remove",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Remove("a")

				return o
			},
			key:  "a",
			want: false,
		},
		{
			name: "exists for key with zero value",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("zero", 0)

				return o
			},
			key:  "zero",
			want: true,
		},
		{
			name: "exists with empty string key",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("", 5)

				return o
			},
			key:  "",
			want: true,
		},
		{
			name: "exists after truncate",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Truncate()

				return o
			},
			key:  "a",
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			testutils.Equal(t, o.Exists(tc.key), tc.want)
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *OrderedMap[string, int]
		want  int
	}{
		{
			name:  "size of empty map",
			setup: func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			want:  0,
		},
		{
			name: "size of single element map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			want: 1,
		},
		{
			name: "size after multiple adds",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			want: 3,
		},
		{
			name: "size after duplicate add",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("a", 2)

				return o
			},
			want: 1,
		},
		{
			name: "size after add and remove",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Remove("a")

				return o
			},
			want: 1,
		},
		{
			name: "size after truncate",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Truncate()

				return o
			},
			want: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			testutils.Equal(t, o.Size(), tc.want)
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		wantSize   int
		wantKeys   []string
		wantValues []int
	}{
		{
			name:       "copy empty map",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			wantSize:   0,
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "copy single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{1},
		},
		{
			name: "copy preserves insertion order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("c", 3)
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			wantSize:   3,
			wantKeys:   []string{"c", "a", "b"},
			wantValues: []int{3, 1, 2},
		},
		{
			name: "copy multiple elements",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("x", 10)
				o.Add("y", 20)
				o.Add("z", 30)
				o.Add("w", 40)
				o.Add("v", 50)

				return o
			},
			wantSize:   5,
			wantKeys:   []string{"x", "y", "z", "w", "v"},
			wantValues: []int{10, 20, 30, 40, 50},
		},
		{
			name: "copy after remove",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)
				o.Remove("b")

				return o
			},
			wantSize:   2,
			wantKeys:   []string{"a", "c"},
			wantValues: []int{1, 3},
		},
		{
			name: "copy is independent from original",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			wantSize:   2,
			wantKeys:   []string{"a", "b"},
			wantValues: []int{1, 2},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			cp := o.Copy()
			testutils.Equal(t, cp.Size(), tc.wantSize)
			keys, values := cp.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}

	t.Run("mutation of copy does not affect original", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		o.Add("b", 2)
		cp := o.Copy()
		cp.Add("c", 3)
		cp.Remove("a")
		testutils.Equal(t, o.Size(), 2)
		testutils.Equal(t, o.Get("a"), 1)
		testutils.Equal(t, o.Exists("c"), false)
	})

	t.Run("mutation of original does not affect copy", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		o.Add("b", 2)
		cp := o.Copy()
		o.Add("d", 4)
		o.Remove("a")
		testutils.Equal(t, cp.Size(), 2)
		testutils.Equal(t, cp.Get("a"), 1)
		testutils.Equal(t, cp.Exists("d"), false)
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *OrderedMap[string, int]
	}{
		{
			name:  "truncate empty map",
			setup: func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
		},
		{
			name: "truncate single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
		},
		{
			name: "truncate multiple elements",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
		},
		{
			name: "truncate already truncated map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Truncate()

				return o
			},
		},
		{
			name: "truncate then re-add",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
		},
		{
			name: "truncate large map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				for i := 0; i < 100; i++ {
					o.Add(string(rune('A'+i)), i)
				}

				return o
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			o.Truncate()
			testutils.Equal(t, o.Size(), 0)
			keys, values := o.GetAll()
			testutils.Equal(t, len(keys), 0)
			testutils.Equal(t, len(values), 0)
		})
	}

	t.Run("add works after truncate", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		o.Add("b", 2)
		o.Truncate()
		o.Add("c", 3)
		testutils.Equal(t, o.Size(), 1)
		testutils.Equal(t, o.Get("c"), 3)
		testutils.Equal(t, o.Exists("a"), false)
	})
}

func TestSetAll(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		newValues  []int
		wantKeys   []string
		wantValues []int
	}{
		{
			name:       "set all on empty map with empty values",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			newValues:  []int{},
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "set all with exact length",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			newValues:  []int{10, 20, 30},
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{10, 20, 30},
		},
		{
			name: "set all with more values than keys uses only matching count",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			newValues:  []int{10, 20, 30, 40},
			wantKeys:   []string{"a", "b"},
			wantValues: []int{10, 20},
		},
		{
			name: "set all with fewer values than keys is no-op",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			newValues:  []int{10},
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{1, 2, 3},
		},
		{
			name: "set all on single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("only", 1)

				return o
			},
			newValues:  []int{99},
			wantKeys:   []string{"only"},
			wantValues: []int{99},
		},
		{
			name: "set all on empty map with non-empty values is no-op",
			setup: func() *OrderedMap[string, int] {
				return &OrderedMap[string, int]{values: map[string]int{}, keys: []string{}}
			},
			newValues:  []int{1, 2, 3},
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "set all with zero values",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 5)
				o.Add("b", 10)

				return o
			},
			newValues:  []int{0, 0},
			wantKeys:   []string{"a", "b"},
			wantValues: []int{0, 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			o.SetAll(tc.newValues)
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestSetKeys(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		keys       []string
		def        int
		wantSize   int
		wantKeys   []string
		wantValues []int
	}{
		{
			name:       "set keys on empty map",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			keys:       []string{"a", "b", "c"},
			def:        0,
			wantSize:   3,
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{0, 0, 0},
		},
		{
			name:       "set keys with empty slice",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			keys:       []string{},
			def:        5,
			wantSize:   0,
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "set keys adds to existing map",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("x", 10)

				return o
			},
			keys:       []string{"a", "b"},
			def:        99,
			wantSize:   3,
			wantKeys:   []string{"x", "a", "b"},
			wantValues: []int{10, 99, 99},
		},
		{
			name: "set keys with duplicates in input",
			setup: func() *OrderedMap[string, int] {
				return &OrderedMap[string, int]{}
			},
			keys:       []string{"a", "a", "a"},
			def:        7,
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{7},
		},
		{
			name: "set keys overwrites existing keys with default",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 100)

				return o
			},
			keys:       []string{"a"},
			def:        0,
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{0},
		},
		{
			name:       "set keys with single key",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			keys:       []string{"only"},
			def:        42,
			wantSize:   1,
			wantKeys:   []string{"only"},
			wantValues: []int{42},
		},
		{
			name: "set keys partially overlapping with existing",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			keys:       []string{"b", "c"},
			def:        0,
			wantSize:   3,
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{1, 0, 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			o.SetKeys(tc.keys, tc.def)
			testutils.Equal(t, o.Size(), tc.wantSize)
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestGetMap(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *OrderedMap[string, int]
		wantMap map[string]int
	}{
		{
			name:    "get map from empty ordered map",
			setup:   func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			wantMap: map[string]int{},
		},
		{
			name: "get map from single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			wantMap: map[string]int{"a": 1},
		},
		{
			name: "get map from multiple elements",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)

				return o
			},
			wantMap: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name: "get map after remove",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Remove("a")

				return o
			},
			wantMap: map[string]int{"b": 2},
		},
		{
			name: "get map after overwrite",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("a", 99)

				return o
			},
			wantMap: map[string]int{"a": 99},
		},
		{
			name: "get map with zero value",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("z", 0)

				return o
			},
			wantMap: map[string]int{"z": 0},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			m := o.GetMap()
			testutils.Equal(t, m, tc.wantMap)
		})
	}

	t.Run("returned map is independent copy", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		m := o.GetMap()
		m["a"] = 999
		m["b"] = 2

		testutils.Equal(t, o.Get("a"), 1)
		testutils.Equal(t, o.Exists("b"), false)
	})
}

func TestGetAll(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		wantKeys   []string
		wantValues []int
	}{
		{
			name:       "get all from empty map",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			wantKeys:   []string{},
			wantValues: []int{},
		},
		{
			name: "get all from single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			wantKeys:   []string{"a"},
			wantValues: []int{1},
		},
		{
			name: "get all preserves insertion order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("z", 26)
				o.Add("a", 1)
				o.Add("m", 13)

				return o
			},
			wantKeys:   []string{"z", "a", "m"},
			wantValues: []int{26, 1, 13},
		},
		{
			name: "get all after removes",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)
				o.Add("d", 4)
				o.Remove("b")
				o.Remove("d")

				return o
			},
			wantKeys:   []string{"a", "c"},
			wantValues: []int{1, 3},
		},
		{
			name: "get all after overwrite",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("a", 100)

				return o
			},
			wantKeys:   []string{"a", "b"},
			wantValues: []int{100, 2},
		},
		{
			name: "get all after truncate and re-add",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Truncate()
				o.Add("x", 10)

				return o
			},
			wantKeys:   []string{"x"},
			wantValues: []int{10},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}

	t.Run("returned keys slice is independent copy", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		o.Add("b", 2)
		keys, _ := o.GetAll()
		keys[0] = "modified"

		testutils.Equal(t, o.Get("a"), 1)
		origKeys, _ := o.GetAll()
		testutils.Equal(t, origKeys[0], "a")
	})
}

func TestIterator(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		bufSize    int
		wantValues []int
	}{
		{
			name:       "iterator on empty map",
			setup:      func() *OrderedMap[string, int] { return &OrderedMap[string, int]{} },
			bufSize:    0,
			wantValues: []int{},
		},
		{
			name: "iterator on single element",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			bufSize:    1,
			wantValues: []int{1},
		},
		{
			name: "iterator preserves insertion order",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("c", 3)
				o.Add("a", 1)
				o.Add("b", 2)

				return o
			},
			bufSize:    3,
			wantValues: []int{3, 1, 2},
		},
		{
			name: "iterator with buffer size zero",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 10)
				o.Add("b", 20)

				return o
			},
			bufSize:    0,
			wantValues: []int{10, 20},
		},
		{
			name: "iterator with buffer larger than size",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			bufSize:    100,
			wantValues: []int{1},
		},
		{
			name: "iterator after remove",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)
				o.Remove("b")

				return o
			},
			bufSize:    2,
			wantValues: []int{1, 3},
		},
		{
			name: "iterator channel closes after all values consumed",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("x", 42)

				return o
			},
			bufSize:    1,
			wantValues: []int{42},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			ch := o.Iterator(tc.bufSize)

			var got []int
			for v := range ch {
				got = append(got, v)
			}

			if len(tc.wantValues) == 0 {
				testutils.Equal(t, len(got), 0)
			} else {
				testutils.Equal(t, got, tc.wantValues)
			}
		})
	}

	t.Run("iterator is snapshot and not affected by later mutations", func(t *testing.T) {
		o := &OrderedMap[string, int]{}
		o.Add("a", 1)
		o.Add("b", 2)
		o.Add("c", 3)
		ch := o.Iterator(0)
		o.Add("d", 4)
		o.Remove("a")

		var got []int
		for v := range ch {
			got = append(got, v)
		}

		testutils.Equal(t, got, []int{1, 2, 3})
	})
}

func TestIteratorSnapshotWithConcurrentMutation(t *testing.T) {
	cases := []struct {
		name    string
		bufSize int
		mutate  func(*OrderedMap[string, int])
		want    []int
	}{
		{
			name:    "add and remove while draining iterator",
			bufSize: 0,
			mutate: func(o *OrderedMap[string, int]) {
				o.Add("d", 4)
				o.Remove("a")
				o.Add("b", 20)
			},
			want: []int{1, 2, 3},
		},
		{
			name:    "truncate while draining iterator",
			bufSize: 0,
			mutate: func(o *OrderedMap[string, int]) {
				o.Truncate()
				o.Add("x", 100)
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := &OrderedMap[string, int]{}
			o.Add("a", 1)
			o.Add("b", 2)
			o.Add("c", 3)
			ch := o.Iterator(tc.bufSize)

			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()

				tc.mutate(o)
			}()

			var got []int
			for value := range ch {
				got = append(got, value)
			}

			wg.Wait()
			testutils.Equal(t, got, tc.want)
		})
	}
}

func TestIntKeys(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[int, string]
		wantKeys   []int
		wantValues []string
	}{
		{
			name: "integer keys preserve order",
			setup: func() *OrderedMap[int, string] {
				o := &OrderedMap[int, string]{}
				o.Add(3, "three")
				o.Add(1, "one")
				o.Add(2, "two")

				return o
			},
			wantKeys:   []int{3, 1, 2},
			wantValues: []string{"three", "one", "two"},
		},
		{
			name: "integer key zero",
			setup: func() *OrderedMap[int, string] {
				o := &OrderedMap[int, string]{}
				o.Add(0, "zero")
				o.Add(-1, "neg")

				return o
			},
			wantKeys:   []int{0, -1},
			wantValues: []string{"zero", "neg"},
		},
		{
			name: "remove with integer keys",
			setup: func() *OrderedMap[int, string] {
				o := &OrderedMap[int, string]{}
				o.Add(1, "a")
				o.Add(2, "b")
				o.Add(3, "c")
				o.Remove(2)

				return o
			},
			wantKeys:   []int{1, 3},
			wantValues: []string{"a", "c"},
		},
		{
			name: "negative keys",
			setup: func() *OrderedMap[int, string] {
				o := &OrderedMap[int, string]{}
				o.Add(-5, "neg5")
				o.Add(-3, "neg3")
				o.Add(-1, "neg1")

				return o
			},
			wantKeys:   []int{-5, -3, -1},
			wantValues: []string{"neg5", "neg3", "neg1"},
		},
		{
			name: "overwrite integer key",
			setup: func() *OrderedMap[int, string] {
				o := &OrderedMap[int, string]{}
				o.Add(1, "old")
				o.Add(1, "new")

				return o
			},
			wantKeys:   []int{1},
			wantValues: []string{"new"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent adds do not panic", func(t *testing.T) {
		o := &OrderedMap[int, int]{}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				o.Add(n, n*10)
			}(i)
		}

		wg.Wait()
		testutils.Equal(t, o.Size(), 100)
	})

	t.Run("concurrent reads and writes do not panic", func(t *testing.T) {
		o := &OrderedMap[int, int]{}
		for i := 0; i < 50; i++ {
			o.Add(i, i)
		}

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(3)

			go func(n int) {
				defer wg.Done()

				o.Add(n+50, n)
			}(i)
			go func(n int) {
				defer wg.Done()

				o.Get(n)
			}(i)
			go func(n int) {
				defer wg.Done()

				o.Exists(n)
			}(i)
		}

		wg.Wait()
		testutils.Equal(t, o.Size(), 100)
	})

	t.Run("concurrent iterator does not panic", func(_ *testing.T) {
		o := &OrderedMap[int, int]{}
		for i := 0; i < 20; i++ {
			o.Add(i, i)
		}

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				ch := o.Iterator(5)
				for value := range ch {
					_ = value
				}
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent copy does not panic", func(_ *testing.T) {
		o := &OrderedMap[int, int]{}
		for i := 0; i < 20; i++ {
			o.Add(i, i)
		}

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				cp := o.Copy()
				_ = cp.Size()
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent removes do not panic", func(t *testing.T) {
		o := &OrderedMap[int, int]{}
		for i := 0; i < 100; i++ {
			o.Add(i, i)
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func(n int) {
				defer wg.Done()

				o.Remove(n)
			}(i)
		}

		wg.Wait()
		testutils.Equal(t, o.Size(), 0)
	})
}

func TestGetMapIndependence(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *OrderedMap[string, int]
		mutate     func(m map[string]int)
		wantGetA   int
		wantExistB bool
	}{
		{
			name: "modifying returned map does not affect original",
			setup: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)

				return o
			},
			mutate:     func(m map[string]int) { m["a"] = 999; m["b"] = 2 },
			wantGetA:   1,
			wantExistB: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.setup()
			m := o.GetMap()
			tc.mutate(m)
			testutils.Equal(t, o.Get("a"), tc.wantGetA)
			testutils.Equal(t, o.Exists("b"), tc.wantExistB)
		})
	}
}

func TestOrderedMapSequentialOperations(t *testing.T) {
	tests := []struct {
		name       string
		ops        func() *OrderedMap[string, int]
		wantSize   int
		wantKeys   []string
		wantValues []int
	}{
		{
			name: "add remove add same key",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Remove("a")
				o.Add("a", 2)

				return o
			},
			wantSize:   1,
			wantKeys:   []string{"a"},
			wantValues: []int{2},
		},
		{
			name: "setkeys then setall",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.SetKeys([]string{"a", "b", "c"}, 0)
				o.SetAll([]int{10, 20, 30})

				return o
			},
			wantSize:   3,
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{10, 20, 30},
		},
		{
			name: "copy then truncate original",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				cp := o.Copy()
				o.Truncate()

				return cp
			},
			wantSize:   2,
			wantKeys:   []string{"a", "b"},
			wantValues: []int{1, 2},
		},
		{
			name: "multiple removes leave correct order",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)
				o.Add("d", 4)
				o.Add("e", 5)
				o.Remove("b")
				o.Remove("d")

				return o
			},
			wantSize:   3,
			wantKeys:   []string{"a", "c", "e"},
			wantValues: []int{1, 3, 5},
		},
		{
			name: "add after setall preserves new values",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.SetAll([]int{10, 20})
				o.Add("c", 30)

				return o
			},
			wantSize:   3,
			wantKeys:   []string{"a", "b", "c"},
			wantValues: []int{10, 20, 30},
		},
		{
			name: "remove all elements one by one",
			ops: func() *OrderedMap[string, int] {
				o := &OrderedMap[string, int]{}
				o.Add("a", 1)
				o.Add("b", 2)
				o.Add("c", 3)
				o.Remove("a")
				o.Remove("b")
				o.Remove("c")

				return o
			},
			wantSize:   0,
			wantKeys:   []string{},
			wantValues: []int{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := tc.ops()
			testutils.Equal(t, o.Size(), tc.wantSize)
			keys, values := o.GetAll()
			testutils.Equal(t, keys, tc.wantKeys)
			testutils.Equal(t, values, tc.wantValues)
		})
	}
}

func TestConcurrentAddsSorted(t *testing.T) {
	t.Run("all concurrent adds are present in map", func(t *testing.T) {
		o := &OrderedMap[int, int]{}

		var wg sync.WaitGroup

		n := 200
		for i := 0; i < n; i++ {
			wg.Add(1)

			go func(v int) {
				defer wg.Done()

				o.Add(v, v*2)
			}(i)
		}

		wg.Wait()

		testutils.Equal(t, o.Size(), n)
		keys, values := o.GetAll()
		testutils.Equal(t, len(keys), n)
		testutils.Equal(t, len(values), n)

		sort.Ints(keys)

		for i := 0; i < n; i++ {
			testutils.Equal(t, keys[i], i)
			testutils.Equal(t, o.Get(i), i*2)
		}
	})
}
