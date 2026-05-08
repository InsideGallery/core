package linkedlist

import (
	"reflect"
	"testing"
)

func TestListOperations(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "insert before and after mark",
			run: func(t *testing.T) {
				t.Helper()

				list := New[string]()
				mark := list.PushBack("b")
				if list.InsertBefore("a", mark) == nil {
					t.Fatal("insert before returned nil")
				}

				if list.InsertAfter("c", mark) == nil {
					t.Fatal("insert after returned nil")
				}

				if got := listValues(list); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
					t.Fatalf("values = %v", got)
				}

				foreign := New[string]().PushBack("foreign")
				if list.InsertBefore("x", foreign) != nil {
					t.Fatal("insert before foreign mark should return nil")
				}

				if list.InsertAfter("x", foreign) != nil {
					t.Fatal("insert after foreign mark should return nil")
				}
			},
		},
		{
			name: "move operations reorder elements",
			run: func(t *testing.T) {
				t.Helper()

				list := New[string]()
				a := list.PushBack("a")
				b := list.PushBack("b")
				c := list.PushBack("c")

				list.MoveToFront(c)
				if got := listValues(list); !reflect.DeepEqual(got, []string{"c", "a", "b"}) {
					t.Fatalf("move front = %v", got)
				}

				list.MoveToBack(c)
				if got := listValues(list); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
					t.Fatalf("move back = %v", got)
				}

				list.MoveBefore(c, b)
				if got := listValues(list); !reflect.DeepEqual(got, []string{"a", "c", "b"}) {
					t.Fatalf("move before = %v", got)
				}

				list.MoveAfter(a, b)
				if got := listValues(list); !reflect.DeepEqual(got, []string{"c", "b", "a"}) {
					t.Fatalf("move after = %v", got)
				}
			},
		},
		{
			name: "remove and list copies",
			run: func(t *testing.T) {
				t.Helper()

				list := New[string]()
				a := list.PushBack("a")
				list.PushBack("b")
				list.PushBack("c")

				if got := list.Remove(a); got != "a" {
					t.Fatalf("removed = %q, want a", got)
				}

				if list.Len() != 2 {
					t.Fatalf("len = %d, want 2", list.Len())
				}

				other := New[string]()
				other.PushBack("x")
				other.PushBack("y")

				list.PushBackList(other)
				if got := listValues(list); !reflect.DeepEqual(got, []string{"b", "c", "x", "y"}) {
					t.Fatalf("push back list = %v", got)
				}

				target := New[string]()
				target.PushBack("z")
				target.PushFrontList(other)
				if got := listValues(target); !reflect.DeepEqual(got, []string{"x", "y", "z"}) {
					t.Fatalf("push front list = %v", got)
				}
			},
		},
		{
			name: "zero value list lazy initializes",
			run: func(t *testing.T) {
				t.Helper()

				var list List[string]
				element := list.PushFront("first")

				if list.Front() != element || list.Back() != element {
					t.Fatal("front and back should point to inserted element")
				}

				if list.ByID(element.ID()) != element {
					t.Fatal("element not found by id")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func listValues(list *List[string]) []string {
	var values []string

	for element := list.Front(); element != nil && !element.IsEmpty(); element = element.Next() {
		values = append(values, element.Value)
	}

	return values
}
