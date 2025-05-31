package btree

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"

	"github.com/InsideGallery/core/memory/comparator"
)

const (
	MinOrder = 3
)

// Tree holds elements of the B-tree
type Tree[K comparable, V any] struct {
	Root       *Node[K, V]           // Root node
	Comparator comparator.Comparator // Key comparator
	size       int                   // Total number of keys in the tree
	m          int                   // order (maximum number of children)
}

// NewWith instantiates a B-tree with the order (maximum number of children) and a custom key comparator.
func NewWith[K comparable, V any](order int, comparator comparator.Comparator) (*Tree[K, V], error) {
	if order < MinOrder {
		return nil, ErrInvalidOrder
	}

	return &Tree[K, V]{m: order, Comparator: comparator}, nil
}

// NewWithIntComparator instantiates a B-tree with the order (maximum number of children)
// and the IntComparator, i.e. keys are of type int.
func NewWithIntComparator(order int) (*Tree[int, any], error) {
	return NewWith[int, any](order, comparator.IntComparator)
}

// NewWithStringComparator instantiates a B-tree with the order (maximum number of children)
// and the StringComparator, i.e. keys are of type string.
func NewWithStringComparator(order int) (*Tree[string, any], error) {
	return NewWith[string, any](order, comparator.StringComparator)
}

// Put inserts key-value pair node into the tree.
// If key already exists, then its value is updated with the new value.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Put(key K, value V) {
	entry := &Entry[K, V]{Key: key, Value: value}

	if tree.Root == nil {
		tree.Root = &Node[K, V]{Entries: []*Entry[K, V]{entry}, Children: []*Node[K, V]{}}
		tree.size++

		return
	}

	if tree.insert(tree.Root, entry) {
		tree.size++
	}
}

// Get searches the node in the tree by key and returns its value or nil if key is not found in tree.
// Second return parameter is true if key was found, otherwise false.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Get(key K) (value V, found bool) {
	node, index, found := tree.searchRecursively(tree.Root, key)
	if found {
		return node.Entries[index].Value, true
	}

	return
}

// Remove remove the node from the tree by key.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *Tree[K, V]) Remove(key K) {
	node, index, found := tree.searchRecursively(tree.Root, key)
	if found {
		tree.delete(node, index)

		tree.size--
	}
}

// Empty returns true if tree does not contain any nodes
func (tree *Tree[K, V]) Empty() bool {
	return tree.size == 0
}

// Size returns number of nodes in the tree.
func (tree *Tree[K, V]) Size() int {
	return tree.size
}

// Keys returns all keys in-order
func (tree *Tree[K, V]) Keys() []K {
	keys := make([]K, tree.size)
	it := tree.Iterator()

	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}

	return keys
}

// Values returns all values in-order based on the key.
func (tree *Tree[K, V]) Values() []any {
	values := make([]any, tree.size)

	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}

	return values
}

// Clear removes all nodes from the tree.
func (tree *Tree[K, V]) Clear() {
	tree.Root = nil
	tree.size = 0
}

// Height returns the height of the tree.
func (tree *Tree[K, V]) Height() int {
	return tree.Root.height()
}

// Left returns the left-most (min) node or nil if tree is empty.
func (tree *Tree[K, V]) Left() *Node[K, V] {
	return tree.left(tree.Root)
}

// LeftKey returns the left-most (min) key or nil if tree is empty.
func (tree *Tree[K, V]) LeftKey() any {
	if left := tree.Left(); left != nil {
		return left.Entries[0].Key
	}

	return nil
}

// LeftValue returns the left-most value or nil if tree is empty.
func (tree *Tree[K, V]) LeftValue() any {
	if left := tree.Left(); left != nil {
		return left.Entries[0].Value
	}

	return nil
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *Tree[K, V]) Right() *Node[K, V] {
	return tree.right(tree.Root)
}

// RightKey returns the right-most (max) key or nil if tree is empty.
func (tree *Tree[K, V]) RightKey() any {
	if right := tree.Right(); right != nil {
		return right.Entries[len(right.Entries)-1].Key
	}

	return nil
}

// RightValue returns the right-most value or nil if tree is empty.
func (tree *Tree[K, V]) RightValue() any {
	if right := tree.Right(); right != nil {
		return right.Entries[len(right.Entries)-1].Value
	}

	return nil
}

// String returns a string representation of container (for debugging purposes)
func (tree *Tree[K, V]) String() string {
	var buffer bytes.Buffer
	if _, err := buffer.WriteString("BTree\n"); err != nil {
		slog.Info("write to output", "err", err)
	}

	if !tree.Empty() {
		tree.output(&buffer, tree.Root, 0)
	}

	return buffer.String()
}

func (tree *Tree[K, V]) output(buffer *bytes.Buffer, node *Node[K, V], level int) {
	for e := 0; e < len(node.Entries)+1; e++ {
		if e < len(node.Children) {
			tree.output(buffer, node.Children[e], level+1)
		}

		if e < len(node.Entries) {
			if _, err := buffer.WriteString(strings.Repeat("    ", level)); err != nil {
				slog.Info("write to output", "err", err)
			}

			if _, err := buffer.WriteString(fmt.Sprintf("%v", node.Entries[e].Key) + "\n"); err != nil {
				slog.Info("write to output", "err", err)
			}
		}
	}
}

func (tree *Tree[K, V]) isLeaf(node *Node[K, V]) bool {
	return len(node.Children) == 0
}

func (tree *Tree[K, V]) shouldSplit(node *Node[K, V]) bool {
	return len(node.Entries) > tree.maxEntries()
}

func (tree *Tree[K, V]) maxChildren() int {
	return tree.m
}

func (tree *Tree[K, V]) minChildren() int {
	return (tree.m + 1) / 2 // nolint:mnd
}

func (tree *Tree[K, V]) maxEntries() int {
	return tree.maxChildren() - 1
}

func (tree *Tree[K, V]) minEntries() int {
	return tree.minChildren() - 1
}

func (tree *Tree[K, V]) middle() int {
	return (tree.m - 1) / 2 // nolint:mnd
}

// search searches only within the single node among its entries
func (tree *Tree[K, V]) search(node *Node[K, V], key any) (index int, found bool) {
	low, high := 0, len(node.Entries)-1
	var mid int

	for low <= high {
		mid = (high + low) / 2 // nolint:mnd
		compare := tree.Comparator(key, node.Entries[mid].Key)

		switch {
		case compare > 0:
			low = mid + 1
		case compare < 0:
			high = mid - 1
		default:
			return mid, true
		}
	}

	return low, false
}

// searchRecursively searches recursively down the tree starting at the startNode
func (tree *Tree[K, V]) searchRecursively(startNode *Node[K, V], key any) (node *Node[K, V], index int, found bool) {
	if tree.Empty() {
		return nil, -1, false
	}

	node = startNode

	for {
		index, found = tree.search(node, key)
		if found {
			return node, index, true
		}

		if tree.isLeaf(node) {
			return nil, -1, false
		}
		node = node.Children[index]
	}
}

func (tree *Tree[K, V]) insert(node *Node[K, V], entry *Entry[K, V]) (inserted bool) {
	if tree.isLeaf(node) {
		return tree.insertIntoLeaf(node, entry)
	}

	return tree.insertIntoInternal(node, entry)
}

func (tree *Tree[K, V]) insertIntoLeaf(node *Node[K, V], entry *Entry[K, V]) (inserted bool) {
	insertPosition, found := tree.search(node, entry.Key)
	if found {
		node.Entries[insertPosition] = entry
		return false
	}
	// Insert entry's key in the middle of the node
	node.Entries = append(node.Entries, nil)
	copy(node.Entries[insertPosition+1:], node.Entries[insertPosition:])
	node.Entries[insertPosition] = entry
	tree.split(node)

	return true
}

func (tree *Tree[K, V]) insertIntoInternal(node *Node[K, V], entry *Entry[K, V]) (inserted bool) {
	insertPosition, found := tree.search(node, entry.Key)
	if found {
		node.Entries[insertPosition] = entry
		return false
	}

	return tree.insert(node.Children[insertPosition], entry)
}

func (tree *Tree[K, V]) split(node *Node[K, V]) {
	if !tree.shouldSplit(node) {
		return
	}

	if node == tree.Root {
		tree.splitRoot()
		return
	}

	tree.splitNonRoot(node)
}

func (tree *Tree[K, V]) splitNonRoot(node *Node[K, V]) {
	middle := tree.middle()
	parent := node.Parent

	left := &Node[K, V]{Entries: append([]*Entry[K, V](nil), node.Entries[:middle]...), Parent: parent}
	right := &Node[K, V]{Entries: append([]*Entry[K, V](nil), node.Entries[middle+1:]...), Parent: parent}

	// Move children from the node to be split into left and right nodes
	if !tree.isLeaf(node) {
		left.Children = append([]*Node[K, V](nil), node.Children[:middle+1]...)
		right.Children = append([]*Node[K, V](nil), node.Children[middle+1:]...)
		tree.setParent(left.Children, left)
		tree.setParent(right.Children, right)
	}

	insertPosition, _ := tree.search(parent, node.Entries[middle].Key)

	// Insert middle key into parent
	parent.Entries = append(parent.Entries, nil)
	copy(parent.Entries[insertPosition+1:], parent.Entries[insertPosition:])
	parent.Entries[insertPosition] = node.Entries[middle]

	// Set child left of inserted key in parent to the created left node
	parent.Children[insertPosition] = left

	// Set child right of inserted key in parent to the created right node
	parent.Children = append(parent.Children, nil)
	copy(parent.Children[insertPosition+2:], parent.Children[insertPosition+1:])
	parent.Children[insertPosition+1] = right

	tree.split(parent)
}

func (tree *Tree[K, V]) splitRoot() {
	middle := tree.middle()

	left := &Node[K, V]{Entries: append([]*Entry[K, V](nil), tree.Root.Entries[:middle]...)}
	right := &Node[K, V]{Entries: append([]*Entry[K, V](nil), tree.Root.Entries[middle+1:]...)}

	// Move children from the node to be split into left and right nodes
	if !tree.isLeaf(tree.Root) {
		left.Children = append([]*Node[K, V](nil), tree.Root.Children[:middle+1]...)
		right.Children = append([]*Node[K, V](nil), tree.Root.Children[middle+1:]...)
		tree.setParent(left.Children, left)
		tree.setParent(right.Children, right)
	}

	// Root is a node with one entry and two children (left and right)
	newRoot := &Node[K, V]{
		Entries:  []*Entry[K, V]{tree.Root.Entries[middle]},
		Children: []*Node[K, V]{left, right},
	}

	left.Parent = newRoot
	right.Parent = newRoot
	tree.Root = newRoot
}

func (tree *Tree[K, V]) setParent(nodes []*Node[K, V], parent *Node[K, V]) {
	for _, node := range nodes {
		node.Parent = parent
	}
}

func (tree *Tree[K, V]) left(node *Node[K, V]) *Node[K, V] {
	if tree.Empty() {
		return nil
	}
	current := node

	for {
		if tree.isLeaf(current) {
			return current
		}

		current = current.Children[0]
	}
}

func (tree *Tree[K, V]) right(node *Node[K, V]) *Node[K, V] {
	if tree.Empty() {
		return nil
	}

	current := node

	for {
		if tree.isLeaf(current) {
			return current
		}
		current = current.Children[len(current.Children)-1]
	}
}

// leftSibling returns the node's left sibling and child index (in parent) if it exists, otherwise (nil,-1)
// key is any of keys in node (could even be deleted).
func (tree *Tree[K, V]) leftSibling(node *Node[K, V], key any) (*Node[K, V], int) {
	if node.Parent != nil {
		index, _ := tree.search(node.Parent, key)
		index--

		if index >= 0 && index < len(node.Parent.Children) {
			return node.Parent.Children[index], index
		}
	}

	return nil, -1
}

// rightSibling returns the node's right sibling and child index (in parent) if it exists, otherwise (nil,-1)
// key is any of keys in node (could even be deleted).
func (tree *Tree[K, V]) rightSibling(node *Node[K, V], key any) (*Node[K, V], int) {
	if node.Parent != nil {
		index, _ := tree.search(node.Parent, key)
		index++

		if index < len(node.Parent.Children) {
			return node.Parent.Children[index], index
		}
	}

	return nil, -1
}

// delete deletes an entry in node at entries' index
// ref.: https://en.wikipedia.org/wiki/B-tree#Deletion
func (tree *Tree[K, V]) delete(node *Node[K, V], index int) {
	// deleting from a leaf node
	if tree.isLeaf(node) {
		deletedKey := node.Entries[index].Key
		tree.deleteEntry(node, index)
		tree.rebalance(node, deletedKey)

		if len(tree.Root.Entries) == 0 {
			tree.Root = nil
		}

		return
	}

	// deleting from an internal node
	leftLargestNode := tree.right(node.Children[index]) // largest node in the left sub-tree (assumed to exist)
	leftLargestEntryIndex := len(leftLargestNode.Entries) - 1
	node.Entries[index] = leftLargestNode.Entries[leftLargestEntryIndex]
	deletedKey := leftLargestNode.Entries[leftLargestEntryIndex].Key
	tree.deleteEntry(leftLargestNode, leftLargestEntryIndex)
	tree.rebalance(leftLargestNode, deletedKey)
}

// rebalance rebalances the tree after deletion if necessary and returns true, otherwise false.
// Note that we first delete the entry and then call rebalance, thus the passed deleted key as reference.
func (tree *Tree[K, V]) rebalance(node *Node[K, V], deletedKey any) {
	// check if rebalancing is needed
	if node == nil || len(node.Entries) >= tree.minEntries() {
		return
	}

	// try to borrow from left sibling
	leftSibling, leftSiblingIndex := tree.leftSibling(node, deletedKey)
	if leftSibling != nil && len(leftSibling.Entries) > tree.minEntries() {
		// rotate right
		// prepend parent's separator entry to node's entries
		node.Entries = append([]*Entry[K, V]{node.Parent.Entries[leftSiblingIndex]}, node.Entries...)
		node.Parent.Entries[leftSiblingIndex] = leftSibling.Entries[len(leftSibling.Entries)-1]
		tree.deleteEntry(leftSibling, len(leftSibling.Entries)-1)

		if !tree.isLeaf(leftSibling) {
			leftSiblingRightMostChild := leftSibling.Children[len(leftSibling.Children)-1]
			leftSiblingRightMostChild.Parent = node
			node.Children = append([]*Node[K, V]{leftSiblingRightMostChild}, node.Children...)
			tree.deleteChild(leftSibling, len(leftSibling.Children)-1)
		}

		return
	}

	// try to borrow from right sibling
	rightSibling, rightSiblingIndex := tree.rightSibling(node, deletedKey)
	if rightSibling != nil && len(rightSibling.Entries) > tree.minEntries() {
		// rotate left
		// append parent's separator entry to node's entries
		node.Entries = append(node.Entries, node.Parent.Entries[rightSiblingIndex-1])
		node.Parent.Entries[rightSiblingIndex-1] = rightSibling.Entries[0]
		tree.deleteEntry(rightSibling, 0)

		if !tree.isLeaf(rightSibling) {
			rightSiblingLeftMostChild := rightSibling.Children[0]
			rightSiblingLeftMostChild.Parent = node
			node.Children = append(node.Children, rightSiblingLeftMostChild)

			tree.deleteChild(rightSibling, 0)
		}

		return
	}

	// merge with siblings
	if rightSibling != nil {
		// merge with right sibling
		node.Entries = append(node.Entries, node.Parent.Entries[rightSiblingIndex-1])
		node.Entries = append(node.Entries, rightSibling.Entries...)
		deletedKey = node.Parent.Entries[rightSiblingIndex-1].Key
		tree.deleteEntry(node.Parent, rightSiblingIndex-1)
		tree.appendChildren(node.Parent.Children[rightSiblingIndex], node)
		tree.deleteChild(node.Parent, rightSiblingIndex)
	} else if leftSibling != nil {
		// merge with left sibling
		entries := append([]*Entry[K, V](nil), leftSibling.Entries...)
		entries = append(entries, node.Parent.Entries[leftSiblingIndex])
		node.Entries = append(entries, node.Entries...)
		deletedKey = node.Parent.Entries[leftSiblingIndex].Key
		tree.deleteEntry(node.Parent, leftSiblingIndex)
		tree.prependChildren(node.Parent.Children[leftSiblingIndex], node)
		tree.deleteChild(node.Parent, leftSiblingIndex)
	}

	// make the merged node the root if its parent was the root and the root is empty
	if node.Parent == tree.Root && len(tree.Root.Entries) == 0 {
		tree.Root = node
		node.Parent = nil

		return
	}

	// parent might underflow, so try to rebalance if necessary
	tree.rebalance(node.Parent, deletedKey)
}

func (tree *Tree[K, V]) prependChildren(fromNode *Node[K, V], toNode *Node[K, V]) {
	children := append([]*Node[K, V](nil), fromNode.Children...)
	toNode.Children = append(children, toNode.Children...)
	tree.setParent(fromNode.Children, toNode)
}

func (tree *Tree[K, V]) appendChildren(fromNode *Node[K, V], toNode *Node[K, V]) {
	toNode.Children = append(toNode.Children, fromNode.Children...)
	tree.setParent(fromNode.Children, toNode)
}

func (tree *Tree[K, V]) deleteEntry(node *Node[K, V], index int) {
	copy(node.Entries[index:], node.Entries[index+1:])
	node.Entries[len(node.Entries)-1] = nil
	node.Entries = node.Entries[:len(node.Entries)-1]
}

func (tree *Tree[K, V]) deleteChild(node *Node[K, V], index int) {
	if index >= len(node.Children) {
		return
	}

	copy(node.Children[index:], node.Children[index+1:])
	node.Children[len(node.Children)-1] = nil
	node.Children = node.Children[:len(node.Children)-1]
}
