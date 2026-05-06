//go:build unit
// +build unit

package stack

import (
	"testing"
)

func TestToSlice(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() *Stack[int]
		expected []int
	}{
		{
			name: "empty stack returns nil slice",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expected: nil,
		},
		{
			name: "single element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(42)
				return s
			},
			expected: []int{42},
		},
		{
			name: "multiple elements preserves order",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Push(3)
				return s
			},
			expected: []int{1, 2, 3},
		},
		{
			name: "after Set",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{10, 20, 30})
				return s
			},
			expected: []int{10, 20, 30},
		},
		{
			name: "after Pop removes last",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{1, 2, 3})
				s.Pop()
				return s
			},
			expected: []int{1, 2},
		},
		{
			name: "returns reference to internal slice",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{5, 6})
				return s
			},
			expected: []int{5, 6},
		},
		{
			name: "zero value stack",
			setup: func() *Stack[int] {
				var s Stack[int]
				return &s
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.ToSlice()
			if tc.expected == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(tc.expected))
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Fatalf("index %d: got %d, want %d", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	cases := []struct {
		name        string
		initial     []int
		setValue     []int
		expectedLen int
		expectedVal []int
	}{
		{
			name:        "set on empty stack",
			initial:     nil,
			setValue:     []int{1, 2, 3},
			expectedLen: 3,
			expectedVal: []int{1, 2, 3},
		},
		{
			name:        "set replaces existing items",
			initial:     []int{10, 20},
			setValue:     []int{30, 40, 50},
			expectedLen: 3,
			expectedVal: []int{30, 40, 50},
		},
		{
			name:        "set with nil clears stack",
			initial:     []int{1, 2},
			setValue:     nil,
			expectedLen: 0,
			expectedVal: nil,
		},
		{
			name:        "set with empty slice",
			initial:     []int{1, 2, 3},
			setValue:     []int{},
			expectedLen: 0,
			expectedVal: []int{},
		},
		{
			name:        "set single element",
			initial:     nil,
			setValue:     []int{99},
			expectedLen: 1,
			expectedVal: []int{99},
		},
		{
			name:        "set with large slice",
			initial:     []int{1},
			setValue:     make([]int, 1000),
			expectedLen: 1000,
			expectedVal: make([]int, 1000),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			if tc.initial != nil {
				s.Set(tc.initial)
			}
			s.Set(tc.setValue)
			if s.Len() != tc.expectedLen {
				t.Fatalf("Len: got %d, want %d", s.Len(), tc.expectedLen)
			}
			got := s.ToSlice()
			if tc.expectedVal == nil {
				if got != nil {
					t.Fatalf("expected nil slice, got %v", got)
				}
				return
			}
			if len(got) != len(tc.expectedVal) {
				t.Fatalf("slice length mismatch: got %d, want %d", len(got), len(tc.expectedVal))
			}
			for i := range tc.expectedVal {
				if got[i] != tc.expectedVal[i] {
					t.Fatalf("index %d: got %d, want %d", i, got[i], tc.expectedVal[i])
				}
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() *Stack[int]
		expected bool
	}{
		{
			name: "zero value stack is empty",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expected: true,
		},
		{
			name: "after push not empty",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				return s
			},
			expected: false,
		},
		{
			name: "after push and pop is empty",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Pop()
				return s
			},
			expected: true,
		},
		{
			name: "after set with empty slice",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{})
				return s
			},
			expected: true,
		},
		{
			name: "after set with nil",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set(nil)
				return s
			},
			expected: true,
		},
		{
			name: "multiple elements not empty",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Push(3)
				return s
			},
			expected: false,
		},
		{
			name: "after PopLeft drains single element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.PopLeft()
				return s
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.IsEmpty()
			if got != tc.expected {
				t.Fatalf("IsEmpty: got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestPeek(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() *Stack[int]
		expected int
	}{
		{
			name: "empty stack returns zero value",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expected: 0,
		},
		{
			name: "single element returns that element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(42)
				return s
			},
			expected: 42,
		},
		{
			name: "multiple elements returns last pushed",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Push(3)
				return s
			},
			expected: 3,
		},
		{
			name: "peek does not remove element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(10)
				s.Peek()
				return s
			},
			expected: 10,
		},
		{
			name: "peek after pop returns new top",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Pop()
				return s
			},
			expected: 1,
		},
		{
			name: "peek after PushLeft returns rightmost",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(5)
				s.PushLeft(1)
				return s
			},
			expected: 5,
		},
		{
			name: "peek on nil-items stack returns zero value",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set(nil)
				return s
			},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.Peek()
			if got != tc.expected {
				t.Fatalf("Peek: got %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestPeekString(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() *Stack[string]
		expected string
	}{
		{
			name: "empty string stack returns empty string",
			setup: func() *Stack[string] {
				return &Stack[string]{}
			},
			expected: "",
		},
		{
			name: "returns last pushed string",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Push("hello")
				s.Push("world")
				return s
			},
			expected: "world",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.Peek()
			if got != tc.expected {
				t.Fatalf("Peek: got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	cases := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "empty stack",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "single element",
			input:    []int{1},
			expected: []int{1},
		},
		{
			name:     "two elements",
			input:    []int{1, 2},
			expected: []int{2, 1},
		},
		{
			name:     "multiple elements",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{5, 4, 3, 2, 1},
		},
		{
			name:     "odd number of elements",
			input:    []int{10, 20, 30},
			expected: []int{30, 20, 10},
		},
		{
			name:     "double reverse restores original",
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "all same elements",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			s.Set(tc.input)
			if tc.name == "double reverse restores original" {
				s.Reverse().Reverse()
			} else {
				result := s.Reverse()
				if result != &s {
					t.Fatal("Reverse should return pointer to same stack")
				}
			}
			got := s.ToSlice()
			if len(got) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(tc.expected))
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Fatalf("index %d: got %d, want %d", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestReverseReturnsSelf(t *testing.T) {
	cases := []struct {
		name  string
		input []int
	}{
		{
			name:  "returns self pointer on non-empty",
			input: []int{1, 2, 3},
		},
		{
			name:  "returns self pointer on empty",
			input: []int{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			s.Set(tc.input)
			result := s.Reverse()
			if result != &s {
				t.Fatal("Reverse must return pointer to the receiver")
			}
		})
	}
}

func TestPush(t *testing.T) {
	cases := []struct {
		name        string
		pushValues  []int
		expectedLen int
		expectedTop int
	}{
		{
			name:        "push to empty stack",
			pushValues:  []int{1},
			expectedLen: 1,
			expectedTop: 1,
		},
		{
			name:        "push multiple values",
			pushValues:  []int{1, 2, 3},
			expectedLen: 3,
			expectedTop: 3,
		},
		{
			name:        "push zero value",
			pushValues:  []int{0},
			expectedLen: 1,
			expectedTop: 0,
		},
		{
			name:        "push negative values",
			pushValues:  []int{-1, -2, -3},
			expectedLen: 3,
			expectedTop: -3,
		},
		{
			name:        "push single large value",
			pushValues:  []int{999999},
			expectedLen: 1,
			expectedTop: 999999,
		},
		{
			name:        "push many values",
			pushValues:  func() []int { v := make([]int, 100); for i := range v { v[i] = i }; return v }(),
			expectedLen: 100,
			expectedTop: 99,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			for _, v := range tc.pushValues {
				s.Push(v)
			}
			if s.Len() != tc.expectedLen {
				t.Fatalf("Len: got %d, want %d", s.Len(), tc.expectedLen)
			}
			if s.Peek() != tc.expectedTop {
				t.Fatalf("Peek: got %d, want %d", s.Peek(), tc.expectedTop)
			}
		})
	}
}

func TestPushLeft(t *testing.T) {
	cases := []struct {
		name          string
		initial       []int
		pushLeftValue int
		expectedSlice []int
	}{
		{
			name:          "push left on empty stack",
			initial:       nil,
			pushLeftValue: 1,
			expectedSlice: []int{1},
		},
		{
			name:          "push left prepends to existing",
			initial:       []int{2, 3},
			pushLeftValue: 1,
			expectedSlice: []int{1, 2, 3},
		},
		{
			name:          "push left on single element",
			initial:       []int{5},
			pushLeftValue: 4,
			expectedSlice: []int{4, 5},
		},
		{
			name:          "push left zero value",
			initial:       []int{1, 2},
			pushLeftValue: 0,
			expectedSlice: []int{0, 1, 2},
		},
		{
			name:          "push left negative value",
			initial:       []int{1},
			pushLeftValue: -10,
			expectedSlice: []int{-10, 1},
		},
		{
			name:          "push left does not affect Peek (top/right)",
			initial:       []int{10, 20},
			pushLeftValue: 5,
			expectedSlice: []int{5, 10, 20},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			if tc.initial != nil {
				s.Set(tc.initial)
			}
			s.PushLeft(tc.pushLeftValue)
			got := s.ToSlice()
			if len(got) != len(tc.expectedSlice) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(tc.expectedSlice))
			}
			for i := range tc.expectedSlice {
				if got[i] != tc.expectedSlice[i] {
					t.Fatalf("index %d: got %d, want %d", i, got[i], tc.expectedSlice[i])
				}
			}
		})
	}
}

func TestPop(t *testing.T) {
	cases := []struct {
		name           string
		setup          func() *Stack[int]
		expectedValue  int
		expectedLenAfter int
	}{
		{
			name: "pop from empty stack returns zero value",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expectedValue:    0,
			expectedLenAfter: 0,
		},
		{
			name: "pop single element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(42)
				return s
			},
			expectedValue:    42,
			expectedLenAfter: 0,
		},
		{
			name: "pop returns last pushed",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Push(3)
				return s
			},
			expectedValue:    3,
			expectedLenAfter: 2,
		},
		{
			name: "pop after set",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{10, 20, 30})
				return s
			},
			expectedValue:    30,
			expectedLenAfter: 2,
		},
		{
			name: "double pop on empty returns zero both times",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Pop()
				return s
			},
			expectedValue:    0,
			expectedLenAfter: 0,
		},
		{
			name: "pop all elements one by one",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Pop()
				return s
			},
			expectedValue:    1,
			expectedLenAfter: 0,
		},
		{
			name: "pop on nil-set stack",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set(nil)
				return s
			},
			expectedValue:    0,
			expectedLenAfter: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.Pop()
			if got != tc.expectedValue {
				t.Fatalf("Pop: got %d, want %d", got, tc.expectedValue)
			}
			if s.Len() != tc.expectedLenAfter {
				t.Fatalf("Len after Pop: got %d, want %d", s.Len(), tc.expectedLenAfter)
			}
		})
	}
}

func TestPopLeft(t *testing.T) {
	cases := []struct {
		name             string
		setup            func() *Stack[string]
		expectedValue    string
		expectedLenAfter int
	}{
		{
			name: "pop left from empty stack returns zero value",
			setup: func() *Stack[string] {
				return &Stack[string]{}
			},
			expectedValue:    "",
			expectedLenAfter: 0,
		},
		{
			name: "pop left single element",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Push("only")
				return s
			},
			expectedValue:    "only",
			expectedLenAfter: 0,
		},
		{
			name: "pop left returns first element",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Set([]string{"first", "second", "third"})
				return s
			},
			expectedValue:    "first",
			expectedLenAfter: 2,
		},
		{
			name: "pop left after PushLeft returns the pushed value",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Push("b")
				s.PushLeft("a")
				return s
			},
			expectedValue:    "a",
			expectedLenAfter: 1,
		},
		{
			name: "pop left multiple times drains from front",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Set([]string{"x", "y", "z"})
				s.PopLeft()
				return s
			},
			expectedValue:    "y",
			expectedLenAfter: 1,
		},
		{
			name: "pop left on nil-set stack",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.Set(nil)
				return s
			},
			expectedValue:    "",
			expectedLenAfter: 0,
		},
		{
			name: "pop left twice on empty stays empty",
			setup: func() *Stack[string] {
				s := &Stack[string]{}
				s.PopLeft()
				return s
			},
			expectedValue:    "",
			expectedLenAfter: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.PopLeft()
			if got != tc.expectedValue {
				t.Fatalf("PopLeft: got %q, want %q", got, tc.expectedValue)
			}
			if s.Len() != tc.expectedLenAfter {
				t.Fatalf("Len after PopLeft: got %d, want %d", s.Len(), tc.expectedLenAfter)
			}
		})
	}
}

func TestLen(t *testing.T) {
	cases := []struct {
		name     string
		setup    func() *Stack[int]
		expected int
	}{
		{
			name: "zero value stack has len 0",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expected: 0,
		},
		{
			name: "after one push",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				return s
			},
			expected: 1,
		},
		{
			name: "after push and pop",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Pop()
				return s
			},
			expected: 0,
		},
		{
			name: "after set with 5 elements",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{1, 2, 3, 4, 5})
				return s
			},
			expected: 5,
		},
		{
			name: "after set nil",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set(nil)
				return s
			},
			expected: 0,
		},
		{
			name: "after multiple pushes and pops",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(1)
				s.Push(2)
				s.Push(3)
				s.Pop()
				s.PopLeft()
				return s
			},
			expected: 1,
		},
		{
			name: "after PushLeft increases length",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.PushLeft(1)
				s.PushLeft(2)
				return s
			},
			expected: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.Len()
			if got != tc.expected {
				t.Fatalf("Len: got %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestStackLIFOBehavior(t *testing.T) {
	cases := []struct {
		name       string
		pushValues []int
		popCount   int
		expected   []int
	}{
		{
			name:       "LIFO order with 3 elements",
			pushValues: []int{1, 2, 3},
			popCount:   3,
			expected:   []int{3, 2, 1},
		},
		{
			name:       "LIFO order with 5 elements",
			pushValues: []int{10, 20, 30, 40, 50},
			popCount:   5,
			expected:   []int{50, 40, 30, 20, 10},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			for _, v := range tc.pushValues {
				s.Push(v)
			}
			var results []int
			for i := 0; i < tc.popCount; i++ {
				results = append(results, s.Pop())
			}
			if len(results) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(results), len(tc.expected))
			}
			for i := range tc.expected {
				if results[i] != tc.expected[i] {
					t.Fatalf("pop %d: got %d, want %d", i, results[i], tc.expected[i])
				}
			}
			if !s.IsEmpty() {
				t.Fatal("stack should be empty after popping all elements")
			}
		})
	}
}

func TestStackFIFOBehaviorWithPopLeft(t *testing.T) {
	cases := []struct {
		name       string
		pushValues []int
		popCount   int
		expected   []int
	}{
		{
			name:       "FIFO order with 3 elements",
			pushValues: []int{1, 2, 3},
			popCount:   3,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "FIFO order with 5 elements",
			pushValues: []int{10, 20, 30, 40, 50},
			popCount:   5,
			expected:   []int{10, 20, 30, 40, 50},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			for _, v := range tc.pushValues {
				s.Push(v)
			}
			var results []int
			for i := 0; i < tc.popCount; i++ {
				results = append(results, s.PopLeft())
			}
			if len(results) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(results), len(tc.expected))
			}
			for i := range tc.expected {
				if results[i] != tc.expected[i] {
					t.Fatalf("popLeft %d: got %d, want %d", i, results[i], tc.expected[i])
				}
			}
			if !s.IsEmpty() {
				t.Fatal("stack should be empty after popping all elements")
			}
		})
	}
}

func TestMixedPushPopOperations(t *testing.T) {
	cases := []struct {
		name     string
		ops      func(s *Stack[int])
		expected []int
	}{
		{
			name: "push right then push left interleaved",
			ops: func(s *Stack[int]) {
				s.Push(2)
				s.PushLeft(1)
				s.Push(3)
				s.PushLeft(0)
			},
			expected: []int{0, 1, 2, 3},
		},
		{
			name: "push then pop then push again",
			ops: func(s *Stack[int]) {
				s.Push(1)
				s.Push(2)
				s.Pop()
				s.Push(3)
			},
			expected: []int{1, 3},
		},
		{
			name: "pop left then push left",
			ops: func(s *Stack[int]) {
				s.Push(10)
				s.Push(20)
				s.PopLeft()
				s.PushLeft(5)
			},
			expected: []int{5, 20},
		},
		{
			name: "reverse then push",
			ops: func(s *Stack[int]) {
				s.Push(1)
				s.Push(2)
				s.Push(3)
				s.Reverse()
				s.Push(4)
			},
			expected: []int{3, 2, 1, 4},
		},
		{
			name: "set then push left then pop",
			ops: func(s *Stack[int]) {
				s.Set([]int{2, 3})
				s.PushLeft(1)
				s.Pop()
			},
			expected: []int{1, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			tc.ops(&s)
			got := s.ToSlice()
			if len(got) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(got), len(tc.expected))
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Fatalf("index %d: got %d, want %d", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestStackWithStructType(t *testing.T) {
	type item struct {
		id   int
		name string
	}

	cases := []struct {
		name         string
		ops          func(s *Stack[item])
		expectedLen  int
		expectedPeek item
	}{
		{
			name: "push struct and peek",
			ops: func(s *Stack[item]) {
				s.Push(item{id: 1, name: "first"})
			},
			expectedLen:  1,
			expectedPeek: item{id: 1, name: "first"},
		},
		{
			name: "pop struct from empty returns zero struct",
			ops: func(s *Stack[item]) {
			},
			expectedLen:  0,
			expectedPeek: item{},
		},
		{
			name: "push multiple structs",
			ops: func(s *Stack[item]) {
				s.Push(item{id: 1, name: "a"})
				s.Push(item{id: 2, name: "b"})
			},
			expectedLen:  2,
			expectedPeek: item{id: 2, name: "b"},
		},
		{
			name: "push left struct",
			ops: func(s *Stack[item]) {
				s.Push(item{id: 2, name: "second"})
				s.PushLeft(item{id: 1, name: "first"})
			},
			expectedLen:  2,
			expectedPeek: item{id: 2, name: "second"},
		},
		{
			name: "pop struct returns correct value",
			ops: func(s *Stack[item]) {
				s.Push(item{id: 1, name: "a"})
				s.Push(item{id: 2, name: "b"})
				s.Pop()
			},
			expectedLen:  1,
			expectedPeek: item{id: 1, name: "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[item]
			tc.ops(&s)
			if s.Len() != tc.expectedLen {
				t.Fatalf("Len: got %d, want %d", s.Len(), tc.expectedLen)
			}
			got := s.Peek()
			if got != tc.expectedPeek {
				t.Fatalf("Peek: got %+v, want %+v", got, tc.expectedPeek)
			}
		})
	}
}

func TestPopLeftInt(t *testing.T) {
	cases := []struct {
		name             string
		setup            func() *Stack[int]
		expectedValue    int
		expectedLenAfter int
	}{
		{
			name: "pop left from empty int stack returns 0",
			setup: func() *Stack[int] {
				return &Stack[int]{}
			},
			expectedValue:    0,
			expectedLenAfter: 0,
		},
		{
			name: "pop left single int element",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Push(99)
				return s
			},
			expectedValue:    99,
			expectedLenAfter: 0,
		},
		{
			name: "pop left first of many",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{100, 200, 300})
				return s
			},
			expectedValue:    100,
			expectedLenAfter: 2,
		},
		{
			name: "pop left preserves remaining order",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{1, 2, 3, 4})
				return s
			},
			expectedValue:    1,
			expectedLenAfter: 3,
		},
		{
			name: "pop left on nil-items int stack",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set(nil)
				return s
			},
			expectedValue:    0,
			expectedLenAfter: 0,
		},
		{
			name: "pop left after reverse",
			setup: func() *Stack[int] {
				s := &Stack[int]{}
				s.Set([]int{1, 2, 3})
				s.Reverse()
				return s
			},
			expectedValue:    3,
			expectedLenAfter: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.setup()
			got := s.PopLeft()
			if got != tc.expectedValue {
				t.Fatalf("PopLeft: got %d, want %d", got, tc.expectedValue)
			}
			if s.Len() != tc.expectedLenAfter {
				t.Fatalf("Len after PopLeft: got %d, want %d", s.Len(), tc.expectedLenAfter)
			}
		})
	}
}

func TestPeekDoesNotMutate(t *testing.T) {
	cases := []struct {
		name  string
		input []int
	}{
		{
			name:  "peek on 3 elements preserves all",
			input: []int{1, 2, 3},
		},
		{
			name:  "peek on 1 element preserves it",
			input: []int{42},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s Stack[int]
			s.Set(tc.input)
			lenBefore := s.Len()
			s.Peek()
			s.Peek()
			s.Peek()
			if s.Len() != lenBefore {
				t.Fatalf("Peek mutated stack length: got %d, want %d", s.Len(), lenBefore)
			}
			got := s.ToSlice()
			for i := range tc.input {
				if got[i] != tc.input[i] {
					t.Fatalf("Peek mutated element at index %d: got %d, want %d", i, got[i], tc.input[i])
				}
			}
		})
	}
}
