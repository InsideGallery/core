package linkedlist

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[V any] struct {
	index map[string]*Element[V] // hash-map to search Element by ID
	root  Element[V]             // sentinel list element, only &root, root.prev, and root.next are used

	len int // current list length excluding (this) sentinel element
}

func (l *List[V]) Init() *List[V] {
	l.index = map[string]*Element[V]{}

	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0

	return l
}

func New[V any]() *List[V] { return new(List[V]).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[V]) Len() int { return l.len }

func (l *List[V]) Front() *Element[V] {
	if l.len == 0 {
		return nil
	}

	return l.root.next
}

func (l *List[V]) Back() *Element[V] {
	if l.len == 0 {
		return nil
	}

	return l.root.prev
}

func (l *List[V]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *List[V]) insert(e, at *Element[V]) *Element[V] {
	var id string
	if entity, ok := l.ValueToAny(e.Value).(Entity); ok {
		id = entity.ID()
	} else {
		id = l.NextID()
	}

	e.id = id

	// Remove previous value
	if _, exists := l.index[e.id]; exists {
		l.Remove(l.ByID(e.id))
	}

	l.index[e.id] = e

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++

	return e
}

func (l *List[V]) insertValue(v V, at *Element[V]) *Element[V] {
	return l.insert(&Element[V]{Value: v}, at)
}

func (l *List[V]) remove(e *Element[V]) {
	delete(l.index, e.id)

	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
}

func (l *List[V]) move(e, at *Element[V]) {
	if e == at {
		return
	}

	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

func (l *List[V]) Remove(e *Element[V]) V {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}

	return e.Value
}

func (l *List[V]) PushFront(v V) *Element[V] {
	l.lazyInit()

	return l.insertValue(v, &l.root)
}

func (l *List[V]) PushBack(v V) *Element[V] {
	l.lazyInit()

	return l.insertValue(v, l.root.prev)
}

func (l *List[V]) InsertBefore(v V, mark *Element[V]) *Element[V] {
	if mark.list != l {
		return nil
	}

	return l.insertValue(v, mark.prev)
}

func (l *List[V]) InsertAfter(v V, mark *Element[V]) *Element[V] {
	if mark.list != l {
		return nil
	}

	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
func (l *List[V]) MoveToFront(e *Element[V]) {
	if e.list != l || l.root.next == e {
		return
	}

	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
func (l *List[V]) MoveToBack(e *Element[V]) {
	if e.list != l || l.root.prev == e {
		return
	}

	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
func (l *List[V]) MoveBefore(e, mark *Element[V]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}

	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
func (l *List[V]) MoveAfter(e, mark *Element[V]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}

	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
func (l *List[V]) PushBackList(other *List[V]) {
	l.lazyInit()

	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
func (l *List[V]) PushFrontList(other *List[V]) {
	l.lazyInit()

	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}

func (l *List[V]) ByID(id string) *Element[V] {
	l.lazyInit()
	return l.index[id]
}

func (l *List[V]) List() (result []V) {
	l.lazyInit()

	front := l.Back()
	for !front.IsEmpty() {
		result = append(result, front.Value)
		front = front.Prev()
	}

	return result
}

func (l *List[V]) Append(elements ...V) {
	l.lazyInit()

	for _, v := range elements {
		l.PushFront(v)
	}
}

func (l *List[V]) ValueToAny(v V) any {
	return v
}
