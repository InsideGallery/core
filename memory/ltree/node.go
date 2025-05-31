package ltree

import (
	"errors"
)

var ErrCircuitDependency = errors.New("circuit dependency")

type Executor[K comparable, V any] interface {
	Key() K
	Value() V
	IsAsync() bool
	DependsOn() []K
}

type Entry[K comparable, V any] struct {
	key     K
	value   V
	need    []K  // related nodes
	isAsync bool // does entity contain blocked operation
}

func (e *Entry[K, V]) Key() K {
	return e.key
}

func (e *Entry[K, V]) Value() V {
	return e.value
}

func (e *Entry[K, V]) DependsOn() []K {
	return e.need
}

func (e *Entry[K, V]) IsAsync() bool {
	return e.isAsync
}

func NewEntry[K comparable, V any](key K, value V, need []K, isAsync bool) *Entry[K, V] {
	return &Entry[K, V]{key: key, value: value, need: need, isAsync: isAsync}
}

type Node[K comparable, V any] struct {
	value    Executor[K, V]
	parent   []*Node[K, V]
	children []*Node[K, V]
	level    int
	visited  bool
}

func NewNode[K comparable, V any](v Executor[K, V]) *Node[K, V] {
	return &Node[K, V]{
		value: v,
	}
}

func (n *Node[K, V]) AddParent(v *Node[K, V]) {
	n.parent = append(n.parent, v)
	v.children = append(v.children, n)

	if n.level < v.level+1 {
		n.level = v.level + 1
	}
}

func (n *Node[K, V]) SetVisited(visited bool) {
	n.visited = visited
}

func (n *Node[K, V]) IsVisited() bool {
	return n.visited
}

func (n *Node[K, V]) HasParent(v *Node[K, V]) bool {
	for _, ch := range n.parent {
		if ch == v {
			return true
		}
	}

	return false
}
