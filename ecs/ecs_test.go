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
	handle := InstallDefaultEntityFactory(NewEntityFactory())
	t.Cleanup(func() {
		if err := handle.Close(); err != nil {
			t.Fatalf("close default entity factory: %v", err)
		}
	})

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

func TestEntityFactoryScopedState(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "factories own id generation",
			run: func(t *testing.T) {
				t.Helper()

				first := NewEntityFactory()
				second := NewEntityFactory()

				firstEntity := first.NewBaseEntity()
				secondEntity := second.NewBaseEntity()

				if firstEntity.GetID() != 1 {
					t.Fatalf("first factory id = %d, want 1", firstEntity.GetID())
				}

				if secondEntity.GetID() != 1 {
					t.Fatalf("second factory id = %d, want 1", secondEntity.GetID())
				}
			},
		},
		{
			name: "set id updates owning factory only",
			run: func(t *testing.T) {
				t.Helper()

				factory := NewEntityFactory()
				other := NewEntityFactory()
				entity := factory.NewBaseEntityWithID(0)

				entity.SetID(10)

				if got := factory.NewBaseEntity().GetID(); got != 11 {
					t.Fatalf("factory next id = %d, want 11", got)
				}

				if got := other.NewBaseEntity().GetID(); got != 1 {
					t.Fatalf("other factory next id = %d, want 1", got)
				}
			},
		},
		{
			name: "default factory handle restores previous factory",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultEntityFactory()
				next := NewEntityFactory()
				handle := InstallDefaultEntityFactory(next)

				if got := DefaultEntityFactory(); got != next {
					t.Fatal("default entity factory was not installed")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default handle: %v", err)
				}

				if got := DefaultEntityFactory(); got != previous {
					t.Fatal("default entity factory was not restored")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
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
