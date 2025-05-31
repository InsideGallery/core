package linkedlist

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

type MockEntity struct {
	id string
}

func (m *MockEntity) ID() string {
	return m.id
}

func TestCustomID(t *testing.T) {
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
