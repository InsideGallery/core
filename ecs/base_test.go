package ecs

import "testing"

func TestRegistryScopedState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new registry owns entity ids",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry()

				first := registry.NewBaseEntity()
				second := registry.NewBaseEntity()

				if first.GetID() != 1 {
					t.Fatalf("first id = %d, want 1", first.GetID())
				}

				if second.GetID() != 2 {
					t.Fatalf("second id = %d, want 2", second.GetID())
				}

				if registry.LatestID() != 2 {
					t.Fatalf("latest id = %d, want 2", registry.LatestID())
				}
			},
		},
		{
			name: "registries isolate id generation",
			run: func(t *testing.T) {
				t.Helper()

				first := NewRegistry()
				second := NewRegistry()

				first.NewBaseEntity()
				first.NewBaseEntity()

				if got := second.NewBaseEntity().GetID(); got != 1 {
					t.Fatalf("second registry id = %d, want 1", got)
				}
			},
		},
		{
			name: "set id advances owning registry only",
			run: func(t *testing.T) {
				t.Helper()

				owner := NewRegistry()
				other := NewRegistry()
				entity := owner.NewBaseEntityWithID(0)

				entity.SetID(25)

				if got := owner.NewBaseEntity().GetID(); got != 26 {
					t.Fatalf("owner next id = %d, want 26", got)
				}

				if got := other.NewBaseEntity().GetID(); got != 1 {
					t.Fatalf("other next id = %d, want 1", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.run(t)
		})
	}
}

func TestDefaultRegistryCompatibility(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "package helpers delegate to default registry",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry()
				handle := InstallDefaultEntityFactory(registry)
				t.Cleanup(func() {
					if err := handle.Close(); err != nil {
						t.Fatalf("close default entity factory: %v", err)
					}
				})

				if Default != registry {
					t.Fatal("default registry was not installed")
				}

				entity := NewBaseEntity()

				if entity.GetID() != 1 {
					t.Fatalf("entity id = %d, want 1", entity.GetID())
				}

				if registry.LatestID() != 1 {
					t.Fatalf("registry latest id = %d, want 1", registry.LatestID())
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
