package linkedlist

import (
	"slices"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

const linkedListLargeCount = 512

type MockEntity struct {
	id string
}

func (m *MockEntity) ID() string {
	return m.id
}

func TestCustomID(t *testing.T) {
	t.Parallel()

	l := New[*MockEntity]()
	test0 := l.PushFront(&MockEntity{id: "test1"})
	l.PushFront(&MockEntity{id: "test2"})
	test1 := l.PushFront(&MockEntity{id: "test2"}) // Replace existing
	test2 := l.PushFront(&MockEntity{id: "test3"})

	testutils.Equal(t, l.ByID(test0.ID()).Root().Value, &MockEntity{id: "test3"})
	testutils.Equal(t, l.ByID(test1.ID()).Root().Value, &MockEntity{id: "test3"})
	testutils.Equal(t, l.ByID(test2.ID()).Root().Value, &MockEntity{id: "test3"})
	testutils.Equal(t, l.ByID(test2.ID()).Root().Next().Value, &MockEntity{id: "test2"})

	testutils.Equal(t, l.ByID(test2.ID()).Value, &MockEntity{id: "test3"})
	testutils.Equal(t, l.ByID(test2.ID()).Next().Value, &MockEntity{id: "test2"})
	testutils.Equal(t, l.ByID(test2.ID()).Next().Next().Value, &MockEntity{id: "test1"})
}

func TestListCopy(t *testing.T) {
	t.Parallel()

	l := New[string]()
	l.PushFront("test")
	test1 := l.PushFront("test1")
	test2 := l.PushFront("test2")
	test3 := l.PushBack("test3")
	l.PushBack("test4")

	testutils.Equal(t, l.List(), []string{"test4", "test3", "test", "test1", "test2"})
	testutils.Equal(t, l.Front().Next().Value, "test1")
	testutils.Equal(t, l.Back().Prev().Value, "test3")

	l2 := New[string]()
	l2.Append(l.List()...) // copy
	testutils.Equal(t, l2.List(), []string{"test4", "test3", "test", "test1", "test2"})
	testutils.Equal(t, l2.Front().Next().Value, "test1")
	testutils.Equal(t, l2.Back().Prev().Value, "test3")

	testutils.Equal(t, l.ByID(test2.ID()).Value, "test2")
	testutils.Equal(t, l.ByID(test2.ID()).Next().Value, "test1")

	testutils.Equal(t, l.ByID(test1.ID()).Value, "test1")
	testutils.Equal(t, l.ByID(test1.ID()).Next().Value, "test")
	testutils.Equal(t, l.ByID(test1.ID()).Prev().Value, "test2")

	testutils.Equal(t, l.ByID(test3.ID()).Value, "test3")
	testutils.Equal(t, l.ByID(test3.ID()).Next().Value, "test4")
	testutils.Equal(t, l.ByID(test3.ID()).Prev().Value, "test")
}

func TestListBoundaryConditions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "empty list has no endpoints",
			run: func(t *testing.T) {
				t.Helper()

				list := New[string]()

				if list.Len() != 0 {
					t.Fatalf("len = %d, want 0", list.Len())
				}

				if list.Front() != nil {
					t.Fatal("front should be nil")
				}

				if list.Back() != nil {
					t.Fatal("back should be nil")
				}

				if list.ByID("missing") != nil {
					t.Fatal("missing id should return nil")
				}
			},
		},
		{
			name: "single element is front and back",
			run: func(t *testing.T) {
				t.Helper()

				list := New[string]()
				element := list.PushBack("only")

				if list.Len() != 1 {
					t.Fatalf("len = %d, want 1", list.Len())
				}

				if list.Front() != element || list.Back() != element {
					t.Fatal("single element should be both front and back")
				}

				if got := list.Remove(element); got != "only" {
					t.Fatalf("removed = %q, want only", got)
				}

				if list.Len() != 0 {
					t.Fatalf("len after remove = %d, want 0", list.Len())
				}
			},
		},
		{
			name: "duplicate entity id replaces prior element",
			run: func(t *testing.T) {
				t.Helper()

				list := New[*MockEntity]()
				first := list.PushBack(&MockEntity{id: "shared"})
				second := list.PushBack(&MockEntity{id: "shared"})

				if list.Len() != 1 {
					t.Fatalf("len = %d, want 1", list.Len())
				}

				if list.ByID(first.ID()) != second {
					t.Fatal("duplicate id should point to replacement element")
				}

				if list.Front() != second || list.Back() != second {
					t.Fatal("replacement should be the only endpoint")
				}
			},
		},
		{
			name: "large append preserves traversal order",
			run: func(t *testing.T) {
				t.Helper()

				list := New[int]()
				elements := make([]*Element[int], 0, linkedListLargeCount)

				for value := 0; value < linkedListLargeCount; value++ {
					elements = append(elements, list.PushBack(value))
				}

				if list.Len() != linkedListLargeCount {
					t.Fatalf("len = %d, want %d", list.Len(), linkedListLargeCount)
				}

				if list.Front().Value != 0 {
					t.Fatalf("front = %d, want 0", list.Front().Value)
				}

				if list.Back().Value != linkedListLargeCount-1 {
					t.Fatalf("back = %d, want %d", list.Back().Value, linkedListLargeCount-1)
				}

				for _, element := range elements {
					if list.ByID(element.ID()) != element {
						t.Fatalf("element %q not found by id", element.ID())
					}
				}

				if got := collectListValues(list); !slices.Equal(got, rangeValues(linkedListLargeCount)) {
					t.Fatalf("values = %v, want ordered range", got)
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

func collectListValues(list *List[int]) []int {
	values := make([]int, 0, list.Len())

	for element := list.Front(); element != nil && !element.IsEmpty(); element = element.Next() {
		values = append(values, element.Value)
	}

	return values
}

func rangeValues(count int) []int {
	values := make([]int, count)
	for value := range count {
		values[value] = value
	}

	return values
}
