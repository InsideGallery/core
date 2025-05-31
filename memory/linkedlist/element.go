package linkedlist

type Entity interface {
	ID() string
}

type Element[V any] struct {
	// The value stored with this element.
	Value V

	next *Element[V]
	prev *Element[V]

	// The list to which this element belongs.
	list *List[V]

	id string // auto-assigned ID
}

func (e *Element[V]) Next() *Element[V] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}

	return &Element[V]{}
}

func (e *Element[V]) Prev() *Element[V] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}

	return &Element[V]{}
}

func (e *Element[V]) Root() *Element[V] {
	return e.list.root.next
}

func (e *Element[V]) IsEmpty() bool {
	return e.id == ""
}

func (e *Element[V]) ID() string {
	return e.id
}
