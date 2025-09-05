package btree

import (
	"fmt"
	"log/slog"
	"testing"
)

func TestBTreeGet1(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, "a")
	tree.Put(2, "b")
	tree.Put(3, "c")
	tree.Put(4, "d")
	tree.Put(5, "e")
	tree.Put(6, "f")
	tree.Put(7, "g")

	v, e := tree.Get(4)
	slog.Info("Test tree get", "v", v, "e", e)

	tests := [][]interface{}{
		{0, nil, false},
		{1, "a", true},
		{2, "b", true},
		{3, "c", true},
		{4, "d", true},
		{5, "e", true},
		{6, "f", true},
		{7, "g", true},
		{8, nil, false},
	}

	for _, test := range tests {
		if value, found := tree.Get(test[0].(int)); value != test[1] || found != test[2] {
			t.Errorf("Got %v,%v expected %v,%v", value, found, test[1], test[2])
		}
	}
}

func TestBTreeGet2(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(7, "g")
	tree.Put(9, "i")
	tree.Put(10, "j")
	tree.Put(6, "f")
	tree.Put(3, "c")
	tree.Put(4, "d")
	tree.Put(5, "e")
	tree.Put(8, "h")
	tree.Put(2, "b")
	tree.Put(1, "a")

	tests := [][]interface{}{
		{0, nil, false},
		{1, "a", true},
		{2, "b", true},
		{3, "c", true},
		{4, "d", true},
		{5, "e", true},
		{6, "f", true},
		{7, "g", true},
		{8, "h", true},
		{9, "i", true},
		{10, "j", true},
		{11, nil, false},
	}

	for _, test := range tests {
		if value, found := tree.Get(test[0].(int)); value != test[1] || found != test[2] {
			t.Errorf("Got %v,%v expected %v,%v", value, found, test[1], test[2])
		}
	}
}

func TestBTreePut1(t *testing.T) {
	// https://upload.wikimedia.org/wikipedia/commons/3/33/B_tree_insertion_example.png
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	assertValidTree(t, tree, 0)

	tree.Put(1, 0)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{1}, false)

	tree.Put(2, 1)
	assertValidTree(t, tree, 2)
	assertValidTreeNode(t, tree.Root, 2, 0, []int{1, 2}, false)

	tree.Put(3, 2)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{2}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{3}, true)

	tree.Put(4, 2)
	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{2}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 2, 0, []int{3, 4}, true)

	tree.Put(5, 2)
	assertValidTree(t, tree, 5)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{2, 4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 0, []int{5}, true)

	tree.Put(6, 2)
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{2, 4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 2, 0, []int{5, 6}, true)

	tree.Put(7, 2)
	assertValidTree(t, tree, 7)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{6}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{7}, true)
}

func TestBTreePut2(t *testing.T) {
	tree, err := NewWithIntComparator(4)
	if err != nil {
		t.Fatal(err)
	}

	assertValidTree(t, tree, 0)

	tree.Put(0, 0)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{0}, false)

	tree.Put(2, 2)
	assertValidTree(t, tree, 2)
	assertValidTreeNode(t, tree.Root, 2, 0, []int{0, 2}, false)

	tree.Put(1, 1)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 3, 0, []int{0, 1, 2}, false)

	tree.Put(1, 1)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 3, 0, []int{0, 1, 2}, false)

	tree.Put(3, 3)
	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{1}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 2, 0, []int{2, 3}, true)

	tree.Put(4, 4)
	assertValidTree(t, tree, 5)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{1}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 3, 0, []int{2, 3, 4}, true)

	tree.Put(5, 5)
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{1, 3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 2, 0, []int{4, 5}, true)
}

func TestBTreePut3(t *testing.T) {
	// http://www.geeksforgeeks.org/b-tree-set-1-insert-2/
	tree, err := NewWithIntComparator(6)
	if err != nil {
		t.Fatal(err)
	}

	assertValidTree(t, tree, 0)

	tree.Put(10, 0)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{10}, false)

	tree.Put(20, 1)
	assertValidTree(t, tree, 2)
	assertValidTreeNode(t, tree.Root, 2, 0, []int{10, 20}, false)

	tree.Put(30, 2)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 3, 0, []int{10, 20, 30}, false)

	tree.Put(40, 3)
	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 4, 0, []int{10, 20, 30, 40}, false)

	tree.Put(50, 4)
	assertValidTree(t, tree, 5)
	assertValidTreeNode(t, tree.Root, 5, 0, []int{10, 20, 30, 40, 50}, false)

	tree.Put(60, 5)
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{30}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{10, 20}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 3, 0, []int{40, 50, 60}, true)

	tree.Put(70, 6)
	assertValidTree(t, tree, 7)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{30}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{10, 20}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 4, 0, []int{40, 50, 60, 70}, true)

	tree.Put(80, 7)
	assertValidTree(t, tree, 8)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{30}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{10, 20}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 5, 0, []int{40, 50, 60, 70, 80}, true)

	tree.Put(90, 8)
	assertValidTree(t, tree, 9)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{30, 60}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{10, 20}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 2, 0, []int{40, 50}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 3, 0, []int{70, 80, 90}, true)
}

func TestBTreePut4(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	assertValidTree(t, tree, 0)

	tree.Put(6, nil)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{6}, false)

	tree.Put(5, nil)
	assertValidTree(t, tree, 2)
	assertValidTreeNode(t, tree.Root, 2, 0, []int{5, 6}, false)

	tree.Put(4, nil)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{5}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{6}, true)

	tree.Put(3, nil)
	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{5}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{3, 4}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{6}, true)

	tree.Put(2, nil)
	assertValidTree(t, tree, 5)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{3, 5}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 0, []int{6}, true)

	tree.Put(1, nil)
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{3, 5}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{1, 2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 0, []int{6}, true)

	tree.Put(0, nil)
	assertValidTree(t, tree, 7)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{6}, true)

	tree.Put(-1, nil)
	assertValidTree(t, tree, 8)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 2, 0, []int{-1, 0}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{6}, true)

	tree.Put(-2, nil)
	assertValidTree(t, tree, 9)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 3, []int{-1, 1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{-2}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[2], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{6}, true)

	tree.Put(-3, nil)
	assertValidTree(t, tree, 10)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 3, []int{-1, 1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 2, 0, []int{-3, -2}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[2], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{6}, true)

	tree.Put(-4, nil)
	assertValidTree(t, tree, 11)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{-1, 3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{-3}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 2, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{-4}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{-2}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[2].Children[0], 1, 0, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[2].Children[1], 1, 0, []int{6}, true)
}

func TestBTreeRemove1(t *testing.T) {
	// empty
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Remove(1)
	assertValidTree(t, tree, 0)
}

func TestBTreeRemove2(t *testing.T) {
	// leaf node (no underflow)
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)

	tree.Remove(1)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{2}, false)

	tree.Remove(2)
	assertValidTree(t, tree, 0)
}

func TestBTreeRemove3(t *testing.T) {
	// merge with right (underflow)
	{
		tree, err := NewWithIntComparator(3)
		if err != nil {
			t.Fatal(err)
		}

		tree.Put(1, nil)
		tree.Put(2, nil)
		tree.Put(3, nil)

		tree.Remove(1)
		assertValidTree(t, tree, 2)
		assertValidTreeNode(t, tree.Root, 2, 0, []int{2, 3}, false)
	}
	// merge with left (underflow)
	{
		tree, err := NewWithIntComparator(3)
		if err != nil {
			t.Fatal(err)
		}

		tree.Put(1, nil)
		tree.Put(2, nil)
		tree.Put(3, nil)

		tree.Remove(3)
		assertValidTree(t, tree, 2)
		assertValidTreeNode(t, tree.Root, 2, 0, []int{1, 2}, false)
	}
}

func TestBTreeRemove4(t *testing.T) {
	// rotate left (underflow)
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)
	tree.Put(3, nil)
	tree.Put(4, nil)

	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{2}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 2, 0, []int{3, 4}, true)

	tree.Remove(1)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{3}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{4}, true)
}

func TestBTreeRemove5(t *testing.T) {
	// rotate right (underflow)
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)
	tree.Put(3, nil)
	tree.Put(0, nil)

	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{2}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{0, 1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{3}, true)

	tree.Remove(3)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{1}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{0}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{2}, true)
}

func TestBTreeRemove6(t *testing.T) {
	// root height reduction after a series of underflows on right side
	// use simulator: https://www.cs.usfca.edu/~galles/visualization/BTree.html
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)
	tree.Put(3, nil)
	tree.Put(4, nil)
	tree.Put(5, nil)
	tree.Put(6, nil)
	tree.Put(7, nil)

	assertValidTree(t, tree, 7)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{6}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{7}, true)

	tree.Remove(7)
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{2, 4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 2, 0, []int{5, 6}, true)
}

func TestBTreeRemove7(t *testing.T) {
	// root height reduction after a series of underflows on left side
	// use simulator: https://www.cs.usfca.edu/~galles/visualization/BTree.html
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)
	tree.Put(3, nil)
	tree.Put(4, nil)
	tree.Put(5, nil)
	tree.Put(6, nil)
	tree.Put(7, nil)

	assertValidTree(t, tree, 7)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{6}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{7}, true)

	tree.Remove(1) // series of underflows
	assertValidTree(t, tree, 6)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{4, 6}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{2, 3}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 0, []int{7}, true)

	// clear all remaining
	tree.Remove(2)
	assertValidTree(t, tree, 5)
	assertValidTreeNode(t, tree.Root, 2, 3, []int{4, 6}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[2], 1, 0, []int{7}, true)

	tree.Remove(3)
	assertValidTree(t, tree, 4)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{6}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 2, 0, []int{4, 5}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{7}, true)

	tree.Remove(4)
	assertValidTree(t, tree, 3)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{6}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 0, []int{7}, true)

	tree.Remove(5)
	assertValidTree(t, tree, 2)
	assertValidTreeNode(t, tree.Root, 2, 0, []int{6, 7}, false)

	tree.Remove(6)
	assertValidTree(t, tree, 1)
	assertValidTreeNode(t, tree.Root, 1, 0, []int{7}, false)

	tree.Remove(7)
	assertValidTree(t, tree, 0)
}

func TestBTreeRemove8(t *testing.T) {
	// use simulator: https://www.cs.usfca.edu/~galles/visualization/BTree.html
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, nil)
	tree.Put(2, nil)
	tree.Put(3, nil)
	tree.Put(4, nil)
	tree.Put(5, nil)
	tree.Put(6, nil)
	tree.Put(7, nil)
	tree.Put(8, nil)
	tree.Put(9, nil)

	assertValidTree(t, tree, 9)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{4}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{2}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 2, 3, []int{6, 8}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 1, 0, []int{1}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{3}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{7}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[2], 1, 0, []int{9}, true)

	tree.Remove(1)
	assertValidTree(t, tree, 8)
	assertValidTreeNode(t, tree.Root, 1, 2, []int{6}, false)
	assertValidTreeNode(t, tree.Root.Children[0], 1, 2, []int{4}, true)
	assertValidTreeNode(t, tree.Root.Children[1], 1, 2, []int{8}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[0], 2, 0, []int{2, 3}, true)
	assertValidTreeNode(t, tree.Root.Children[0].Children[1], 1, 0, []int{5}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[0], 1, 0, []int{7}, true)
	assertValidTreeNode(t, tree.Root.Children[1].Children[1], 1, 0, []int{9}, true)
}

func TestBTreeRemove9(t *testing.T) {
	const maxElm = 1000

	orders := []int{3, 4, 5, 6, 7, 8, 9, 10, 20, 100, 500, 1000, 5000, 10000}
	for _, order := range orders {
		tree, err := NewWithIntComparator(order)
		if err != nil {
			t.Fatal(err)
		}

		{
			for i := 1; i <= maxElm; i++ {
				tree.Put(i, i)
			}

			assertValidTree(t, tree, maxElm)

			for i := 1; i <= maxElm; i++ {
				if _, found := tree.Get(i); !found {
					t.Errorf("Not found %v", i)
				}
			}

			for i := 1; i <= maxElm; i++ {
				tree.Remove(i)
			}

			assertValidTree(t, tree, 0)
		}

		{
			for i := maxElm; i > 0; i-- {
				tree.Put(i, i)
			}

			assertValidTree(t, tree, maxElm)

			for i := maxElm; i > 0; i-- {
				if _, found := tree.Get(i); !found {
					t.Errorf("Not found %v", i)
				}
			}

			for i := maxElm; i > 0; i-- {
				tree.Remove(i)
			}

			assertValidTree(t, tree, 0)
		}
	}
}

func TestBTreeHeight(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	if actualValue, expectedValue := tree.Height(), 0; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(1, 0)

	if actualValue, expectedValue := tree.Height(), 1; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(2, 1)

	if actualValue, expectedValue := tree.Height(), 1; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(3, 2)

	if actualValue, expectedValue := tree.Height(), 2; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(4, 2)

	if actualValue, expectedValue := tree.Height(), 2; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(5, 2)

	if actualValue, expectedValue := tree.Height(), 2; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(6, 2)

	if actualValue, expectedValue := tree.Height(), 2; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Put(7, 2)

	if actualValue, expectedValue := tree.Height(), 3; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	tree.Remove(1)
	tree.Remove(2)
	tree.Remove(3)
	tree.Remove(4)
	tree.Remove(5)
	tree.Remove(6)
	tree.Remove(7)

	if actualValue, expectedValue := tree.Height(), 0; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeLeftAndRight(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	if actualValue := tree.Left(); actualValue != nil {
		t.Errorf("Got %v expected %v", actualValue, nil)
	}

	if actualValue := tree.Right(); actualValue != nil {
		t.Errorf("Got %v expected %v", actualValue, nil)
	}

	tree.Put(1, "a")
	tree.Put(5, "e")
	tree.Put(6, "f")
	tree.Put(7, "g")
	tree.Put(3, "c")
	tree.Put(4, "d")
	tree.Put(1, "x") // overwrite
	tree.Put(2, "b")

	if actualValue, expectedValue := tree.LeftKey(), 1; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if actualValue, expectedValue := tree.LeftValue(), "x"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if actualValue, expectedValue := tree.RightKey(), 7; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if actualValue, expectedValue := tree.RightValue(), "g"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIteratorValuesAndKeys(t *testing.T) {
	tree, err := NewWithIntComparator(4)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(4, "d")
	tree.Put(5, "e")
	tree.Put(6, "f")
	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(7, "g")
	tree.Put(2, "b")
	tree.Put(1, "x") // override
	slog.Info("Get keys", "keys", tree.Keys())

	if actualValue, expectedValue := fmt.Sprintf("%s%s%s%s%s%s%s", tree.Values()...), "xbcdefg"; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if actualValue := tree.Size(); actualValue != 7 {
		t.Errorf("Got %v expected %v", actualValue, 7)
	}
}

func TestBTreeIteratorNextOnEmpty(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	it := tree.Iterator()
	for it.Next() {
		t.Errorf("Shouldn't iterate on empty tree")
	}
}

func TestBTreeIteratorPrevOnEmpty(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	it := tree.Iterator()
	for it.Prev() {
		t.Errorf("Shouldn't iterate on empty tree")
	}
}

func TestBTreeIterator1Next(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(5, "e")
	tree.Put(6, "f")
	tree.Put(7, "g")
	tree.Put(3, "c")
	tree.Put(4, "d")
	tree.Put(1, "x")
	tree.Put(2, "b")
	tree.Put(1, "a") // overwrite
	it := tree.Iterator()

	count := 0
	for it.Next() {
		count++

		key := it.Key()
		switch key {
		case count:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}

	if actualValue, expectedValue := count, tree.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator1Prev(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(5, "e")
	tree.Put(6, "f")
	tree.Put(7, "g")
	tree.Put(3, "c")
	tree.Put(4, "d")
	tree.Put(1, "x")
	tree.Put(2, "b")
	tree.Put(1, "a") // overwrite

	it := tree.Iterator()
	for it.Next() {
	}

	countDown := tree.size

	for it.Prev() {
		key := it.Key()
		switch key {
		case countDown:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}

		countDown--
	}

	if actualValue, expectedValue := countDown, 0; actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator2Next(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")
	it := tree.Iterator()

	count := 0
	for it.Next() {
		count++

		key := it.Key()
		switch key {
		case count:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}

	if actualValue, expectedValue := count, tree.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator2Prev(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")

	it := tree.Iterator()
	for it.Next() {
	}

	countDown := tree.size

	for it.Prev() {
		key := it.Key()
		switch key {
		case countDown:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}

		countDown--
	}

	if actualValue, expectedValue := countDown, 0; actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator3Next(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, "a")
	it := tree.Iterator()

	count := 0
	for it.Next() {
		count++

		key := it.Key()
		switch key {
		case count:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}

	if actualValue, expectedValue := count, tree.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator3Prev(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(1, "a")

	it := tree.Iterator()
	for it.Next() {
	}

	countDown := tree.size

	for it.Prev() {
		key := it.Key()
		switch key {
		case countDown:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := key, countDown; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}

		countDown--
	}

	if actualValue, expectedValue := countDown, 0; actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator4Next(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(13, 5)
	tree.Put(8, 3)
	tree.Put(17, 7)
	tree.Put(1, 1)
	tree.Put(11, 4)
	tree.Put(15, 6)
	tree.Put(25, 9)
	tree.Put(6, 2)
	tree.Put(22, 8)
	tree.Put(27, 10)
	it := tree.Iterator()

	count := 0
	for it.Next() {
		count++

		value := it.Value()
		switch value {
		case count:
			if actualValue, expectedValue := value, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := value, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}

	if actualValue, expectedValue := count, tree.Size(); actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIterator4Prev(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(13, 5)
	tree.Put(8, 3)
	tree.Put(17, 7)
	tree.Put(1, 1)
	tree.Put(11, 4)
	tree.Put(15, 6)
	tree.Put(25, 9)
	tree.Put(6, 2)
	tree.Put(22, 8)
	tree.Put(27, 10)
	it := tree.Iterator()
	count := tree.Size()

	for it.Next() {
	}

	for it.Prev() {
		value := it.Value()
		switch value {
		case count:
			if actualValue, expectedValue := value, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		default:
			if actualValue, expectedValue := value, count; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}

		count--
	}

	if actualValue, expectedValue := count, 0; actualValue != expectedValue {
		t.Errorf("Size different. Got %v expected %v", actualValue, expectedValue)
	}
}

func TestBTreeIteratorBegin(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")
	it := tree.Iterator()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	it.Begin()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	for it.Next() {
	}

	it.Begin()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	it.Next()

	if key, value := it.Key(), it.Value(); key != 1 || value != "a" {
		t.Errorf("Got %v,%v expected %v,%v", key, value, 1, "a")
	}
}

func TestBTreeIteratorEnd(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	it := tree.Iterator()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	it.End()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")

	it.End()

	if it.node != nil {
		t.Errorf("Got %v expected %v", it.node, nil)
	}

	it.Prev()

	if key, value := it.Key(), it.Value(); key != 3 || value != "c" {
		t.Errorf("Got %v,%v expected %v,%v", key, value, 3, "c")
	}
}

func TestBTreeIteratorFirst(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")

	it := tree.Iterator()
	if actualValue, expectedValue := it.First(), true; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if key, value := it.Key(), it.Value(); key != 1 || value != "a" {
		t.Errorf("Got %v,%v expected %v,%v", key, value, 1, "a")
	}
}

func TestBTreeIteratorLast(t *testing.T) {
	tree, err := NewWithIntComparator(3)
	if err != nil {
		t.Fatal(err)
	}

	tree.Put(3, "c")
	tree.Put(1, "a")
	tree.Put(2, "b")

	it := tree.Iterator()
	if actualValue, expectedValue := it.Last(), true; actualValue != expectedValue {
		t.Errorf("Got %v expected %v", actualValue, expectedValue)
	}

	if key, value := it.Key(), it.Value(); key != 3 || value != "c" {
		t.Errorf("Got %v,%v expected %v,%v", key, value, 3, "c")
	}
}

func TestBTree_search(t *testing.T) {
	{
		tree, err := NewWithIntComparator(3)
		if err != nil {
			t.Fatal(err)
		}

		tree.Root = &Node[int, any]{Entries: []*Entry[int, any]{}, Children: make([]*Node[int, any], 0)}

		tests := [][]interface{}{
			{0, 0, false},
		}
		for _, test := range tests {
			index, found := tree.search(tree.Root, test[0])
			if actualValue, expectedValue := index, test[1]; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}

			if actualValue, expectedValue := found, test[2]; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}
	{
		tree, err := NewWithIntComparator(3)
		if err != nil {
			t.Fatal(err)
		}

		tree.Root = &Node[int, any]{Entries: []*Entry[int, any]{{2, 0}, {4, 1}, {6, 2}}, Children: []*Node[int, any]{}}

		tests := [][]interface{}{
			{0, 0, false},
			{1, 0, false},
			{2, 0, true},
			{3, 1, false},
			{4, 1, true},
			{5, 2, false},
			{6, 2, true},
			{7, 3, false},
		}
		for _, test := range tests {
			index, found := tree.search(tree.Root, test[0])
			if actualValue, expectedValue := index, test[1]; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}

			if actualValue, expectedValue := found, test[2]; actualValue != expectedValue {
				t.Errorf("Got %v expected %v", actualValue, expectedValue)
			}
		}
	}
}

func assertValidTree(t *testing.T, tree *Tree[int, any], expectedSize int) {
	if actualValue, expectedValue := tree.size, expectedSize; actualValue != expectedValue {
		t.Errorf("Got %v expected %v for tree size", actualValue, expectedValue)
	}
}

func assertValidTreeNode(t *testing.T, node *Node[int, any], expectedEntries int, expectedChildren int, keys []int, hasParent bool) {
	if actualValue, expectedValue := node.Parent != nil, hasParent; actualValue != expectedValue {
		t.Errorf("Got %v expected %v for hasParent", actualValue, expectedValue)
	}

	if actualValue, expectedValue := len(node.Entries), expectedEntries; actualValue != expectedValue {
		t.Errorf("Got %v expected %v for entries size", actualValue, expectedValue)
	}

	if actualValue, expectedValue := len(node.Children), expectedChildren; actualValue != expectedValue {
		t.Errorf("Got %v expected %v for children size", actualValue, expectedValue)
	}

	for i, key := range keys {
		if actualValue, expectedValue := node.Entries[i].Key, key; actualValue != expectedValue {
			t.Errorf("Got %v expected %v for key", actualValue, expectedValue)
		}
	}
}

func benchmarkGet(b *testing.B, tree *Tree[int, any], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			tree.Get(n)
		}
	}
}

func benchmarkPut(b *testing.B, tree *Tree[int, any], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			tree.Put(n, struct{}{})
		}
	}
}

func benchmarkRemove(b *testing.B, tree *Tree[int, any], size int) {
	for i := 0; i < b.N; i++ {
		for n := 0; n < size; n++ {
			tree.Remove(n)
		}
	}
}

/*
BenchmarkBTreeGet100-12                   271334              4238 ns/op               0 B/op          0 allocs/op
BenchmarkBTreeGet1000-12                   12499            103069 ns/op            5952 B/op        744 allocs/op
BenchmarkBTreeGet10000-12                    908           1229919 ns/op           77955 B/op       9744 allocs/op
BenchmarkBTreeGet100000-12                    81          15505512 ns/op          797957 B/op      99744 allocs/op
BenchmarkBTreePut100-12                   147436             11404 ns/op            3200 B/op        100 allocs/op
BenchmarkBTreePut1000-12                    9090            134715 ns/op           37952 B/op       1744 allocs/op
BenchmarkBTreePut10000-12                    727           1663305 ns/op          397958 B/op      19744 allocs/op
BenchmarkBTreePut100000-12                    62          19119667 ns/op         3997978 B/op     199744 allocs/op
BenchmarkBTreeRemove100-12               1952350               540 ns/op               0 B/op          0 allocs/op
BenchmarkBTreeRemove1000-12                93166             14885 ns/op            5952 B/op        744 allocs/op
BenchmarkBTreeRemove10000-12                6432            178063 ns/op           77952 B/op       9744 allocs/op
BenchmarkBTreeRemove100000-12                610           1824637 ns/op          797960 B/op      99744 allocs/op
*/
func BenchmarkBTreeGet100(b *testing.B) {
	b.StopTimer()

	size := 100

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkGet(b, tree, size)
}

func BenchmarkBTreeGet1000(b *testing.B) {
	b.StopTimer()

	size := 1000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkGet(b, tree, size)
}

func BenchmarkBTreeGet10000(b *testing.B) {
	b.StopTimer()

	size := 10000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkGet(b, tree, size)
}

func BenchmarkBTreeGet100000(b *testing.B) {
	b.StopTimer()

	size := 100000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkGet(b, tree, size)
}

func BenchmarkBTreePut100(b *testing.B) {
	b.StopTimer()

	size := 100

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	benchmarkPut(b, tree, size)
}

func BenchmarkBTreePut1000(b *testing.B) {
	b.StopTimer()

	size := 1000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkPut(b, tree, size)
}

func BenchmarkBTreePut10000(b *testing.B) {
	b.StopTimer()

	size := 10000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkPut(b, tree, size)
}

func BenchmarkBTreePut100000(b *testing.B) {
	b.StopTimer()

	size := 100000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkPut(b, tree, size)
}

func BenchmarkBTreeRemove100(b *testing.B) {
	b.StopTimer()

	size := 100

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkRemove(b, tree, size)
}

func BenchmarkBTreeRemove1000(b *testing.B) {
	b.StopTimer()

	size := 1000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkRemove(b, tree, size)
}

func BenchmarkBTreeRemove10000(b *testing.B) {
	b.StopTimer()

	size := 10000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkRemove(b, tree, size)
}

func BenchmarkBTreeRemove100000(b *testing.B) {
	b.StopTimer()

	size := 100000

	tree, err := NewWithIntComparator(128)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < size; n++ {
		tree.Put(n, struct{}{})
	}

	b.StartTimer()
	benchmarkRemove(b, tree, size)
}
