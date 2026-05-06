package utils

import (
	"strconv"
	"sync"
	"testing"
)

func TestNewSafeMap(t *testing.T) {
	cases := []struct {
		name     string
		input    map[string]int
		wantLen  int
		wantKeys []string
	}{
		{
			name:     "nil map",
			input:    nil,
			wantLen:  0,
			wantKeys: []string{},
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			wantLen:  0,
			wantKeys: []string{},
		},
		{
			name:     "single entry",
			input:    map[string]int{"a": 1},
			wantLen:  1,
			wantKeys: []string{"a"},
		},
		{
			name:     "multiple entries",
			input:    map[string]int{"a": 1, "b": 2, "c": 3},
			wantLen:  3,
			wantKeys: []string{"a", "b", "c"},
		},
		{
			name:     "keys with empty string",
			input:    map[string]int{"": 0},
			wantLen:  1,
			wantKeys: []string{""},
		},
		{
			name:     "large map",
			input:    makeLargeMap(100),
			wantLen:  100,
			wantKeys: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.input)
			got := sm.GetMap()
			if len(got) != tc.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tc.wantLen)
			}
			for _, k := range tc.wantKeys {
				if _, ok := got[k]; !ok {
					t.Errorf("key %q not found", k)
				}
			}
		})
	}
}

func TestNewSafeMap_DoesNotMutateInput(t *testing.T) {
	cases := []struct {
		name  string
		input map[string]int
	}{
		{
			name:  "verify input isolation",
			input: map[string]int{"a": 1, "b": 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := make(map[string]int)
			for k, v := range tc.input {
				original[k] = v
			}
			sm := NewSafeMap(tc.input)
			sm.Set("z", 99)
			if _, ok := tc.input["z"]; ok {
				t.Error("NewSafeMap should copy data, not reference it")
			}
			if len(tc.input) != len(original) {
				t.Error("input map was mutated")
			}
		})
	}
}

func TestSafeMap_Set(t *testing.T) {
	cases := []struct {
		name      string
		setup     map[string]int
		setKey    string
		setValue  int
		wantValue int
		wantOk    bool
	}{
		{
			name:      "set new key on empty map",
			setup:     map[string]int{},
			setKey:    "foo",
			setValue:  42,
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "overwrite existing key",
			setup:     map[string]int{"foo": 1},
			setKey:    "foo",
			setValue:  99,
			wantValue: 99,
			wantOk:    true,
		},
		{
			name:      "set empty string key",
			setup:     map[string]int{},
			setKey:    "",
			setValue:  0,
			wantValue: 0,
			wantOk:    true,
		},
		{
			name:      "set zero value",
			setup:     map[string]int{},
			setKey:    "zero",
			setValue:  0,
			wantValue: 0,
			wantOk:    true,
		},
		{
			name:      "set negative value",
			setup:     map[string]int{},
			setKey:    "neg",
			setValue:  -100,
			wantValue: -100,
			wantOk:    true,
		},
		{
			name:      "set with nil initial map",
			setup:     nil,
			setKey:    "key",
			setValue:  10,
			wantValue: 10,
			wantOk:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			sm.Set(tc.setKey, tc.setValue)
			got, ok := sm.Get(tc.setKey)
			if ok != tc.wantOk {
				t.Fatalf("Get ok = %v, want %v", ok, tc.wantOk)
			}
			if got != tc.wantValue {
				t.Errorf("Get value = %d, want %d", got, tc.wantValue)
			}
		})
	}
}

func TestSafeMap_Get(t *testing.T) {
	cases := []struct {
		name      string
		setup     map[string]int
		getKey    string
		wantValue int
		wantOk    bool
	}{
		{
			name:      "get existing key",
			setup:     map[string]int{"x": 10},
			getKey:    "x",
			wantValue: 10,
			wantOk:    true,
		},
		{
			name:      "get non-existing key",
			setup:     map[string]int{"x": 10},
			getKey:    "y",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "get from empty map",
			setup:     map[string]int{},
			getKey:    "any",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "get empty string key exists",
			setup:     map[string]int{"": 5},
			getKey:    "",
			wantValue: 5,
			wantOk:    true,
		},
		{
			name:      "get empty string key missing",
			setup:     map[string]int{"a": 1},
			getKey:    "",
			wantValue: 0,
			wantOk:    false,
		},
		{
			name:      "get zero value stored",
			setup:     map[string]int{"z": 0},
			getKey:    "z",
			wantValue: 0,
			wantOk:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			got, ok := sm.Get(tc.getKey)
			if ok != tc.wantOk {
				t.Fatalf("Get ok = %v, want %v", ok, tc.wantOk)
			}
			if got != tc.wantValue {
				t.Errorf("Get value = %d, want %d", got, tc.wantValue)
			}
		})
	}
}

func TestSafeMap_Exists(t *testing.T) {
	cases := []struct {
		name  string
		setup map[string]int
		key   string
		want  bool
	}{
		{
			name:  "existing key",
			setup: map[string]int{"a": 1},
			key:   "a",
			want:  true,
		},
		{
			name:  "missing key",
			setup: map[string]int{"a": 1},
			key:   "b",
			want:  false,
		},
		{
			name:  "empty map",
			setup: map[string]int{},
			key:   "a",
			want:  false,
		},
		{
			name:  "empty string key present",
			setup: map[string]int{"": 0},
			key:   "",
			want:  true,
		},
		{
			name:  "empty string key absent",
			setup: map[string]int{"a": 1},
			key:   "",
			want:  false,
		},
		{
			name:  "key with zero value",
			setup: map[string]int{"zero": 0},
			key:   "zero",
			want:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			if got := sm.Exists(tc.key); got != tc.want {
				t.Errorf("Exists(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestSafeMap_Remove(t *testing.T) {
	cases := []struct {
		name        string
		setup       map[string]int
		removeKey   string
		wantExists  bool
		wantLen     int
	}{
		{
			name:       "remove existing key",
			setup:      map[string]int{"a": 1, "b": 2},
			removeKey:  "a",
			wantExists: false,
			wantLen:    1,
		},
		{
			name:       "remove non-existing key",
			setup:      map[string]int{"a": 1},
			removeKey:  "b",
			wantExists: false,
			wantLen:    1,
		},
		{
			name:       "remove from empty map",
			setup:      map[string]int{},
			removeKey:  "a",
			wantExists: false,
			wantLen:    0,
		},
		{
			name:       "remove empty string key",
			setup:      map[string]int{"": 5, "a": 1},
			removeKey:  "",
			wantExists: false,
			wantLen:    1,
		},
		{
			name:       "remove last key",
			setup:      map[string]int{"only": 1},
			removeKey:  "only",
			wantExists: false,
			wantLen:    0,
		},
		{
			name:       "remove then verify other keys intact",
			setup:      map[string]int{"a": 1, "b": 2, "c": 3},
			removeKey:  "b",
			wantExists: false,
			wantLen:    2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			sm.Remove(tc.removeKey)
			if got := sm.Exists(tc.removeKey); got != tc.wantExists {
				t.Errorf("Exists(%q) after Remove = %v, want %v", tc.removeKey, got, tc.wantExists)
			}
			if got := len(sm.GetMap()); got != tc.wantLen {
				t.Errorf("len after Remove = %d, want %d", got, tc.wantLen)
			}
		})
	}
}

func TestSafeMap_GetMap(t *testing.T) {
	cases := []struct {
		name    string
		setup   map[string]int
		wantLen int
	}{
		{
			name:    "empty map",
			setup:   map[string]int{},
			wantLen: 0,
		},
		{
			name:    "nil initial",
			setup:   nil,
			wantLen: 0,
		},
		{
			name:    "single entry",
			setup:   map[string]int{"a": 1},
			wantLen: 1,
		},
		{
			name:    "multiple entries",
			setup:   map[string]int{"a": 1, "b": 2, "c": 3},
			wantLen: 3,
		},
		{
			name:    "entry with empty key",
			setup:   map[string]int{"": 0},
			wantLen: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			got := sm.GetMap()
			if len(got) != tc.wantLen {
				t.Fatalf("GetMap len = %d, want %d", len(got), tc.wantLen)
			}
			for k, v := range tc.setup {
				if gv, ok := got[k]; !ok || gv != v {
					t.Errorf("GetMap[%q] = %d, %v; want %d, true", k, gv, ok, v)
				}
			}
		})
	}
}

func TestSafeMap_GetMap_ReturnsCopy(t *testing.T) {
	cases := []struct {
		name  string
		setup map[string]int
	}{
		{
			name:  "mutation of returned map does not affect SafeMap",
			setup: map[string]int{"a": 1, "b": 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap(tc.setup)
			got := sm.GetMap()
			got["injected"] = 999
			if sm.Exists("injected") {
				t.Error("GetMap should return a copy, not internal reference")
			}
		})
	}
}

func TestSafeMap_ConcurrentAccess(t *testing.T) {
	cases := []struct {
		name       string
		goroutines int
		opsPerG    int
	}{
		{name: "10 goroutines 100 ops", goroutines: 10, opsPerG: 100},
		{name: "50 goroutines 50 ops", goroutines: 50, opsPerG: 50},
		{name: "100 goroutines 10 ops", goroutines: 100, opsPerG: 10},
		{name: "1 goroutine 1000 ops", goroutines: 1, opsPerG: 1000},
		{name: "200 goroutines 5 ops", goroutines: 200, opsPerG: 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap[string, int](nil)
			var wg sync.WaitGroup
			wg.Add(tc.goroutines * 3)

			for g := 0; g < tc.goroutines; g++ {
				go func(id int) {
					defer wg.Done()
					for i := 0; i < tc.opsPerG; i++ {
						sm.Set(strconv.Itoa(id*1000+i), i)
					}
				}(g)

				go func(id int) {
					defer wg.Done()
					for i := 0; i < tc.opsPerG; i++ {
						sm.Get(strconv.Itoa(id*1000 + i))
					}
				}(g)

				go func(id int) {
					defer wg.Done()
					for i := 0; i < tc.opsPerG; i++ {
						sm.Exists(strconv.Itoa(id*1000 + i))
						sm.Remove(strconv.Itoa(id*1000 + i))
						sm.GetMap()
					}
				}(g)
			}

			wg.Wait()
		})
	}
}

func TestSafeMap_SetGetRemoveCycle(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value int
	}{
		{name: "normal key", key: "hello", value: 42},
		{name: "empty key", key: "", value: 0},
		{name: "long key", key: "a_very_long_key_name_that_goes_on_and_on", value: -1},
		{name: "unicode key", key: "clef_\u266B", value: 100},
		{name: "whitespace key", key: "  ", value: 7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewSafeMap[string, int](nil)

			sm.Set(tc.key, tc.value)
			if !sm.Exists(tc.key) {
				t.Fatal("key should exist after Set")
			}

			got, ok := sm.Get(tc.key)
			if !ok || got != tc.value {
				t.Fatalf("Get = (%d, %v), want (%d, true)", got, ok, tc.value)
			}

			sm.Remove(tc.key)
			if sm.Exists(tc.key) {
				t.Fatal("key should not exist after Remove")
			}

			got2, ok2 := sm.Get(tc.key)
			if ok2 {
				t.Fatalf("Get after Remove = (%d, %v), want (0, false)", got2, ok2)
			}
		})
	}
}

func makeLargeMap(n int) map[string]int {
	m := make(map[string]int, n)
	for i := 0; i < n; i++ {
		m[strconv.Itoa(i)] = i
	}
	return m
}
