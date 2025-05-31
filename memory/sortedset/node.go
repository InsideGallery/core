package sortedset

type Node[K comparable, V comparable] struct {
	value    V // associated data
	key      K // key to determine the order of this node in the set
	backward *Node[K, V]
	level    []Level[K, V]
}

func (s *Node[K, V]) Value() V {
	return s.value
}

func (s *Node[K, V]) Key() K {
	return s.key
}

type Level[K comparable, V comparable] struct {
	forward *Node[K, V]
	span    uint64
}
