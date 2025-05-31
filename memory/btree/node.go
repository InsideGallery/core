package btree

import "fmt"

// Node is a single element within the tree
type Node[K comparable, V any] struct {
	Parent   *Node[K, V]
	Entries  []*Entry[K, V] // Contained keys in node
	Children []*Node[K, V]  // Children nodes
}

func (node *Node[K, V]) height() int {
	height := 0
	for ; node != nil; node = node.Children[0] {
		height++

		if len(node.Children) == 0 {
			break
		}
	}

	return height
}

// Entry represents the key-value pair contained within nodes
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

func (entry *Entry[K, V]) String() string {
	return fmt.Sprintf("%v", entry.Key)
}
