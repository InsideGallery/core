package utils

import (
	"sync"
	"testing"
)

func TestNewSafeList(t *testing.T) {
	cases := []struct {
		name    string
		input   []int
		wantLen int
	}{
		{
			name:    "no arguments",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "single element",
			input:   []int{1},
			wantLen: 1,
		},
		{
			name:    "multiple elements",
			input:   []int{1, 2, 3, 4, 5},
			wantLen: 5,
		},
		{
			name:    "zero values",
			input:   []int{0, 0, 0},
			wantLen: 3,
		},
		{
			name:    "negative values",
			input:   []int{-1, -2, -3},
			wantLen: 3,
		},
		{
			name:    "empty explicit slice",
			input:   []int{},
			wantLen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.input == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.input...)
			}

			if got := sl.Count(); got != tc.wantLen {
				t.Fatalf("Count() = %d, want %d", got, tc.wantLen)
			}

			list := sl.List()
			if len(list) != tc.wantLen {
				t.Fatalf("List() len = %d, want %d", len(list), tc.wantLen)
			}

			for i, v := range tc.input {
				if list[i] != v {
					t.Errorf("List()[%d] = %d, want %d", i, list[i], v)
				}
			}
		})
	}
}

func TestSafeList_Add(t *testing.T) {
	cases := []struct {
		name     string
		initial  []int
		addItems []int
		wantList []int
	}{
		{
			name:     "add to empty list",
			initial:  nil,
			addItems: []int{1},
			wantList: []int{1},
		},
		{
			name:     "add multiple to empty list",
			initial:  nil,
			addItems: []int{1, 2, 3},
			wantList: []int{1, 2, 3},
		},
		{
			name:     "add to non-empty list",
			initial:  []int{10, 20},
			addItems: []int{30},
			wantList: []int{10, 20, 30},
		},
		{
			name:     "add zero value",
			initial:  []int{1},
			addItems: []int{0},
			wantList: []int{1, 0},
		},
		{
			name:     "add duplicates",
			initial:  []int{5},
			addItems: []int{5, 5, 5},
			wantList: []int{5, 5, 5, 5},
		},
		{
			name:     "add negative values",
			initial:  nil,
			addItems: []int{-3, -2, -1},
			wantList: []int{-3, -2, -1},
		},
		{
			name:     "add nothing",
			initial:  []int{1, 2},
			addItems: nil,
			wantList: []int{1, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.initial == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			for _, item := range tc.addItems {
				sl.Add(item)
			}

			got := sl.List()
			if len(got) != len(tc.wantList) {
				t.Fatalf("List() len = %d, want %d", len(got), len(tc.wantList))
			}

			for i := range tc.wantList {
				if got[i] != tc.wantList[i] {
					t.Errorf("List()[%d] = %d, want %d", i, got[i], tc.wantList[i])
				}
			}
		})
	}
}

func TestSafeList_List(t *testing.T) {
	cases := []struct {
		name    string
		initial []int
		want    []int
	}{
		{
			name:    "empty list",
			initial: nil,
			want:    []int{},
		},
		{
			name:    "single element",
			initial: []int{42},
			want:    []int{42},
		},
		{
			name:    "preserves order",
			initial: []int{3, 1, 2},
			want:    []int{3, 1, 2},
		},
		{
			name:    "all zeros",
			initial: []int{0, 0, 0},
			want:    []int{0, 0, 0},
		},
		{
			name:    "large list",
			initial: makeSequence(1000),
			want:    makeSequence(1000),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.initial == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			got := sl.List()
			if len(got) != len(tc.want) {
				t.Fatalf("List() len = %d, want %d", len(got), len(tc.want))
			}

			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("List()[%d] = %d, want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestSafeList_ListReturnsCopy(t *testing.T) {
	cases := []struct {
		name    string
		initial []int
	}{
		{name: "mutation of returned slice does not affect SafeList", initial: []int{1, 2, 3}},
		{name: "single element copy", initial: []int{99}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sl := NewSafeList(tc.initial...)
			got := sl.List()
			got[0] = -999

			original := sl.List()
			if original[0] == -999 {
				t.Error("List should return a copy, not the internal slice")
			}
		})
	}
}

func TestSafeList_Reset(t *testing.T) {
	cases := []struct {
		name           string
		initial        []int
		wantReturned   []int
		wantCountAfter int
	}{
		{
			name:           "reset non-empty list",
			initial:        []int{1, 2, 3},
			wantReturned:   []int{1, 2, 3},
			wantCountAfter: 0,
		},
		{
			name:           "reset empty list",
			initial:        nil,
			wantReturned:   []int{},
			wantCountAfter: 0,
		},
		{
			name:           "reset single element",
			initial:        []int{42},
			wantReturned:   []int{42},
			wantCountAfter: 0,
		},
		{
			name:           "reset preserves order in returned",
			initial:        []int{5, 3, 1, 4, 2},
			wantReturned:   []int{5, 3, 1, 4, 2},
			wantCountAfter: 0,
		},
		{
			name:           "reset zeros",
			initial:        []int{0, 0, 0},
			wantReturned:   []int{0, 0, 0},
			wantCountAfter: 0,
		},
		{
			name:           "reset large list",
			initial:        makeSequence(500),
			wantReturned:   makeSequence(500),
			wantCountAfter: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.initial == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			got := sl.Reset()
			if len(got) != len(tc.wantReturned) {
				t.Fatalf("Reset() len = %d, want %d", len(got), len(tc.wantReturned))
			}

			for i := range tc.wantReturned {
				if got[i] != tc.wantReturned[i] {
					t.Errorf("Reset()[%d] = %d, want %d", i, got[i], tc.wantReturned[i])
				}
			}

			if c := sl.Count(); c != tc.wantCountAfter {
				t.Errorf("Count() after Reset = %d, want %d", c, tc.wantCountAfter)
			}
		})
	}
}

func TestSafeList_ResetThenAdd(t *testing.T) {
	cases := []struct {
		name     string
		initial  []int
		addAfter []int
		wantList []int
	}{
		{
			name:     "reset then add",
			initial:  []int{1, 2, 3},
			addAfter: []int{10, 20},
			wantList: []int{10, 20},
		},
		{
			name:     "reset empty then add",
			initial:  nil,
			addAfter: []int{5},
			wantList: []int{5},
		},
		{
			name:     "reset then add nothing",
			initial:  []int{1},
			addAfter: nil,
			wantList: []int{},
		},
		{
			name:     "double reset then add",
			initial:  []int{1, 2},
			addAfter: []int{99},
			wantList: []int{99},
		},
		{
			name:     "reset then add many",
			initial:  []int{1},
			addAfter: makeSequence(100),
			wantList: makeSequence(100),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.initial == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			sl.Reset()

			for _, v := range tc.addAfter {
				sl.Add(v)
			}

			got := sl.List()
			if len(got) != len(tc.wantList) {
				t.Fatalf("List() len = %d, want %d", len(got), len(tc.wantList))
			}

			for i := range tc.wantList {
				if got[i] != tc.wantList[i] {
					t.Errorf("List()[%d] = %d, want %d", i, got[i], tc.wantList[i])
				}
			}
		})
	}
}

func TestSafeList_Count(t *testing.T) {
	cases := []struct {
		name    string
		initial []int
		adds    int
		resets  int
		want    int
	}{
		{
			name:    "empty list",
			initial: nil,
			adds:    0,
			resets:  0,
			want:    0,
		},
		{
			name:    "initial elements only",
			initial: []int{1, 2, 3},
			adds:    0,
			resets:  0,
			want:    3,
		},
		{
			name:    "initial plus adds",
			initial: []int{1},
			adds:    4,
			resets:  0,
			want:    5,
		},
		{
			name:    "after reset",
			initial: []int{1, 2, 3},
			adds:    0,
			resets:  1,
			want:    0,
		},
		{
			name:    "reset then adds",
			initial: []int{1, 2, 3},
			adds:    2,
			resets:  1,
			want:    2,
		},
		{
			name:    "multiple resets",
			initial: []int{1},
			adds:    0,
			resets:  3,
			want:    0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[int]
			if tc.initial == nil {
				sl = NewSafeList[int]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			for i := 0; i < tc.resets; i++ {
				sl.Reset()
			}

			for i := 0; i < tc.adds; i++ {
				sl.Add(i)
			}

			if got := sl.Count(); got != tc.want {
				t.Errorf("Count() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSafeList_ConcurrentAccess(t *testing.T) {
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
		t.Run(tc.name, func(_ *testing.T) {
			sl := NewSafeList[int]()

			var wg sync.WaitGroup
			wg.Add(tc.goroutines * 4)

			for g := 0; g < tc.goroutines; g++ {
				go func() {
					defer wg.Done()

					for i := 0; i < tc.opsPerG; i++ {
						sl.Add(i)
					}
				}()

				go func() {
					defer wg.Done()

					for i := 0; i < tc.opsPerG; i++ {
						sl.List()
					}
				}()

				go func() {
					defer wg.Done()

					for i := 0; i < tc.opsPerG; i++ {
						sl.Count()
					}
				}()

				go func() {
					defer wg.Done()

					for i := 0; i < tc.opsPerG; i++ {
						sl.Reset()
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestSafeList_StringType(t *testing.T) {
	cases := []struct {
		name    string
		initial []string
		add     []string
		want    []string
	}{
		{
			name:    "empty string list",
			initial: nil,
			add:     nil,
			want:    []string{},
		},
		{
			name:    "string list with empty strings",
			initial: []string{"", ""},
			add:     []string{""},
			want:    []string{"", "", ""},
		},
		{
			name:    "mixed strings",
			initial: []string{"hello"},
			add:     []string{"world", "!"},
			want:    []string{"hello", "world", "!"},
		},
		{
			name:    "unicode strings",
			initial: nil,
			add:     []string{"\u0041", "\u00e9", "\u4e16"},
			want:    []string{"\u0041", "\u00e9", "\u4e16"},
		},
		{
			name:    "long strings",
			initial: []string{string(make([]byte, 1000))},
			add:     nil,
			want:    []string{string(make([]byte, 1000))},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sl *SafeList[string]
			if tc.initial == nil {
				sl = NewSafeList[string]()
			} else {
				sl = NewSafeList(tc.initial...)
			}

			for _, s := range tc.add {
				sl.Add(s)
			}

			got := sl.List()
			if len(got) != len(tc.want) {
				t.Fatalf("List() len = %d, want %d", len(got), len(tc.want))
			}

			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("List()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func makeSequence(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}

	return s
}
