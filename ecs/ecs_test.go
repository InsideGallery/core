package ecs

import (
	"context"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

type TestSystem struct {
	components []*TestComponent
}

func (s *TestSystem) Update(_ context.Context) {
	for _, c := range s.components {
		c.Text = "abc"
	}
}

type TestComponent struct {
	Text string
}

type TestEntity struct {
	*TestComponent
}

func TestBaseEntity(t *testing.T) {
	e := NewBaseEntity()
	testutils.Equal(t, e.GetID(), uint64(1))
	testutils.Equal(t, e.GetVersion(), uint64(1))
	e = NewBaseEntity()
	testutils.Equal(t, e.GetID(), uint64(2))
	testutils.Equal(t, e.GetVersion(), uint64(1))
	e.UpVersion()
	testutils.Equal(t, e.GetVersion(), uint64(2))
	e = NewBaseEntityWithID(10)
	testutils.Equal(t, e.GetID(), uint64(10))
	testutils.Equal(t, e.GetVersion(), uint64(1))
	e.UpVersion()
	testutils.Equal(t, e.GetVersion(), uint64(2))
}

func TestECS(t *testing.T) {
	text := &TestComponent{
		Text: "bcd",
	}

	entity := &TestEntity{
		TestComponent: text,
	}

	system := &TestSystem{
		components: []*TestComponent{
			text,
		},
	}
	system.Update(context.Background())

	testutils.Equal(t, entity.Text, "abc")
}
