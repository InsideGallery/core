package registry

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestRegistryAdditionalOperations(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "latest id and indexes",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, int]()
				registry.SetLatestID(41)

				if got := registry.LatestID(); got != 41 {
					t.Fatalf("latest id = %d, want 41", got)
				}

				if got := registry.NextID(); got != 42 {
					t.Fatalf("next id = %d, want 42", got)
				}

				registry.AddIndex(42, "item")
				if got := registry.GetIndex(42); got != "item" {
					t.Fatalf("index = %q, want item", got)
				}

				registry.RemIndex(42)
				if got := registry.GetIndex(42); got != "" {
					t.Fatalf("removed index = %q, want empty", got)
				}
			},
		},
		{
			name: "iterators return requested groups",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, int]()
				mustAdd(t, registry, "a", "1", 1)
				mustAdd(t, registry, "a", "2", 2)
				mustAdd(t, registry, "b", "3", 3)

				if got := collectRegistryValues(registry.Iterator("a", "b")); !reflect.DeepEqual(got, []int{1, 2, 3}) {
					t.Fatalf("iterator = %v", got)
				}

				got := collectRegistryValues(registry.AsyncIterator("a", "b"))
				sort.Ints(got)
				if !reflect.DeepEqual(got, []int{1, 2, 3}) {
					t.Fatalf("async iterator = %v", got)
				}
			},
		},
		{
			name: "tick groups and async tick call tickers",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, any]()
				ticker := &MockEntity{}
				mustAdd(t, registry, "a", "1", any(ticker))

				registry.TickGroups("a")
				registry.AsyncTick("a")

				if ticker.counter != 2 {
					t.Fatalf("ticks = %d, want 2", ticker.counter)
				}
			},
		},
		{
			name: "remove id everywhere removes from each group",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, int]()
				mustAdd(t, registry, "a", "shared", 1)
				mustAdd(t, registry, "b", "shared", 2)

				if err := registry.RemoveIDEverywhere("shared"); err != nil {
					t.Fatalf("remove everywhere: %v", err)
				}

				if _, err := registry.Get("a", "shared"); !errors.Is(err, ErrNotFoundEntity) {
					t.Fatalf("err = %v, want %v", err, ErrNotFoundEntity)
				}

				if registry.Size() != 2 {
					t.Fatalf("size = %d, want 2 groups", registry.Size())
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func mustAdd[G comparable, I comparable, V any](t *testing.T, registry *Registry[G, I, V], group G, id I, value V) {
	t.Helper()

	if err := registry.Add(group, id, value); err != nil {
		t.Fatalf("add: %v", err)
	}
}

func collectRegistryValues[V any](ch chan V) []V {
	var values []V
	for value := range ch {
		values = append(values, value)
	}

	return values
}
