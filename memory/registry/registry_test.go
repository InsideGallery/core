package registry

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

const KeyTemporary = "test"

const (
	registryLargeCount  = 512
	registryReaderCount = 8
)

type MockEntity struct {
	counter int
	id      uint64
	log     []string
	x, y    int
}

func (m *MockEntity) GetID() uint64 {
	return m.id
}

func (m *MockEntity) Tick() {
	m.counter++
}

func (m *MockEntity) Construct() error {
	m.log = append(m.log, "calling construct")
	return nil
}

func (m *MockEntity) Destroy() error {
	m.log = append(m.log, "calling destroy")
	return nil
}

func (m *MockEntity) Coordinates() (int, int) {
	return m.x, m.y
}

func TestRegistry(t *testing.T) {
	r := NewRegistry[string, uint64, any]()
	testcases := map[string]struct {
		id     uint64
		key    string
		data   interface{}
		setErr error
		getErr error
		result interface{}
	}{
		"simple_usage": {
			id:     r.NextID(),
			key:    "string_key",
			data:   "simple string",
			setErr: nil,
			getErr: nil,
			result: "simple string",
		},
		"empty_key": {
			id:     r.NextID(),
			key:    "",
			data:   "simple string",
			setErr: nil,
			getErr: nil,
			result: "simple string",
		},
		"empty_id": {
			id:     0,
			key:    "string_key2",
			data:   "simple string",
			setErr: nil,
			getErr: nil,
			result: "simple string",
		},
		"struct_store": {
			id:  r.NextID(),
			key: "string_key3",
			data: &MockEntity{
				id: 2,
			},
			setErr: nil,
			getErr: nil,
			result: &MockEntity{
				id: 2,
			},
		},
	}

	for name, test := range testcases {
		test := test

		t.Run(name, func(t *testing.T) {
			err := r.Add(test.key, test.id, test.data)
			testutils.Equal(t, err, test.setErr)
			data, err = r.Get(test.key, test.id)
			testutils.Equal(t, err, test.getErr)
			testutils.Equal(t, data, test.data)
		})
	}
}

func TestConstructDestroy(t *testing.T) {
	r := NewRegistry[string, uint64, any]()
	d := &MockEntity{
		id: r.NextID(),
	}

	err := r.Add(KeyTemporary, d.id, d)
	if err != nil {
		t.Fatal(err)
	}

	err = r.Add(KeyTemporary, d.id, d) // UPDATE
	if err != nil {
		t.Fatal(err)
	}

	err = r.Remove(KeyTemporary, d.id)
	if err != nil {
		t.Fatal(err)
	}

	result := strings.Join(d.log, "\n")
	if result != "calling construct\ncalling construct\ncalling destroy" {
		t.Fatalf("Unexpected result: %s", result)
	}
}

func TestAsyncGetNextID(t *testing.T) {
	r := NewRegistry[uint64, string, any]()
	generate := make(chan struct{}, 100)

	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func() {
			for range generate {
				r.NextID()
			}

			wg.Done()
		}()
	}

	for i := 0; i < 100; i++ {
		generate <- struct{}{}
	}

	close(generate)
	wg.Wait()

	id := r.NextID()
	if id != 101 {
		t.Fatalf("Unexpected next id: %d", id)
	}
}

func TestAsyncRegistry(t *testing.T) {
	r := NewRegistry[string, uint64, any]()
	writer := make(chan *MockEntity, 100)

	gorutines := 100

	var wg sync.WaitGroup
	wg.Add(gorutines)

	for i := 0; i < gorutines; i++ {
		go func() {
			for entity := range writer {
				err := r.Add(KeyTemporary, entity.GetID(), entity)
				if err != nil {
					panic(err)
				}
			}

			wg.Done()
		}()
	}

	for i := 0; i < 100; i++ {
		e := &MockEntity{
			id: r.NextID(),
		}
		writer <- e
	}

	close(writer)
	wg.Wait()

	count := len(r.GetGroup(KeyTemporary).GetValues())
	if count != 100 {
		t.Fatalf("Unexpected count: %d", count)
	}

	e := r.SearchOne(KeyTemporary, func(_ interface{}, id interface{}, _ interface{}) bool {
		return id.(uint64) == 1
	})

	entity, ok := e.(*MockEntity)
	if !ok {
		t.Fatalf("Unexpected type: %+v", e)
	}

	result := strings.Join(entity.log, "\n")
	if result != "calling construct" {
		t.Fatalf("Unexpected result: %s", result)
	}

	if entity.GetID() != 1 {
		t.Fatalf("Unexpected id: %d", entity.GetID())
	}
}

func TestRegistryBoundaryConditions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "empty registry lazily creates missing group",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, int]()

				if registry.Size() != 0 {
					t.Fatalf("size = %d, want 0", registry.Size())
				}

				if _, err := registry.Get("missing", "id"); !errors.Is(err, ErrNotFoundEntity) {
					t.Fatalf("err = %v, want %v", err, ErrNotFoundEntity)
				}

				if registry.Size() != 1 {
					t.Fatalf("size after get = %d, want 1", registry.Size())
				}
			},
		},
		{
			name: "single element can be added read and removed",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, int]()
				mustAddRegistryValue(t, registry, "group", "one", 1)

				got, err := registry.Get("group", "one")
				if err != nil {
					t.Fatalf("get: %v", err)
				}

				if got != 1 {
					t.Fatalf("value = %d, want 1", got)
				}

				if err := registry.Remove("group", "one"); err != nil {
					t.Fatalf("remove: %v", err)
				}

				if _, err := registry.Get("group", "one"); !errors.Is(err, ErrNotFoundEntity) {
					t.Fatalf("err = %v, want %v", err, ErrNotFoundEntity)
				}
			},
		},
		{
			name: "duplicate id updates existing value",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, string, string]()
				mustAddRegistryValue(t, registry, "group", "shared", "first")
				mustAddRegistryValue(t, registry, "group", "shared", "second")

				got, err := registry.Get("group", "shared")
				if err != nil {
					t.Fatalf("get: %v", err)
				}

				if got != "second" {
					t.Fatalf("value = %q, want second", got)
				}

				if values := registry.GetValues("group"); len(values) != 1 {
					t.Fatalf("values len = %d, want 1", len(values))
				}
			},
		},
		{
			name: "large id range preserves values and indexes",
			run: func(t *testing.T) {
				t.Helper()

				registry := NewRegistry[string, uint64, uint64]()

				for id := uint64(0); id < registryLargeCount; id++ {
					mustAddRegistryValue(t, registry, "group", id, id)
					registry.AddIndex(id, id)
				}

				if values := registry.GetValues("group"); len(values) != registryLargeCount {
					t.Fatalf("values len = %d, want %d", len(values), registryLargeCount)
				}

				for _, id := range []uint64{0, registryLargeCount / 2, registryLargeCount - 1} {
					got, err := registry.Get("group", id)
					if err != nil {
						t.Fatalf("get %d: %v", id, err)
					}

					if got != id {
						t.Fatalf("value = %d, want %d", got, id)
					}

					if got := registry.GetIndex(id); got != id {
						t.Fatalf("index = %d, want %d", got, id)
					}
				}
			},
		},
	}

	for _, test := range cases {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.run(t)
		})
	}
}

func TestRegistryConcurrentReads(t *testing.T) {
	t.Parallel()

	registry := NewRegistry[string, uint64, uint64]()
	for id := uint64(0); id < registryLargeCount; id++ {
		mustAddRegistryValue(t, registry, "group", id, id)
	}

	errCh := make(chan error, registryReaderCount)

	var wg sync.WaitGroup
	wg.Add(registryReaderCount)

	for range registryReaderCount {
		go func() {
			defer wg.Done()

			for id := uint64(0); id < registryLargeCount; id++ {
				got, err := registry.Get("group", id)
				if err != nil {
					errCh <- fmt.Errorf("get id %d: %w", id, err)
					return
				}

				if got != id {
					errCh <- fmt.Errorf("value = %d, want %d", got, id)
					return
				}
			}

			if values := registry.GetValues("group"); len(values) != registryLargeCount {
				errCh <- fmt.Errorf("values len = %d, want %d", len(values), registryLargeCount)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func mustAddRegistryValue[G comparable, I comparable, V any](
	t *testing.T,
	registry *Registry[G, I, V],
	group G,
	id I,
	value V,
) {
	t.Helper()

	if err := registry.Add(group, id, value); err != nil {
		t.Fatalf("add: %v", err)
	}
}

/*
BenchmarkGetString-4            20000000                63.1 ns/op             0 B/op          0 allocs/op
BenchmarkGetEntity-4            20000000                64.5 ns/op             0 B/op          0 allocs/op
BenchmarkSetString-4             5000000               387 ns/op              83 B/op          0 allocs/op
BenchmarkSetEntity-4             1000000              1455 ns/op             184 B/op          2 allocs/op
BenchmarkSetDeleteString-4       5000000               283 ns/op               0 B/op          0 allocs/op
BenchmarkSetDeleteEntity-4       2000000               653 ns/op             112 B/op          3 allocs/op
*/

// Prevent optimization
var data interface{}

func BenchmarkGetString(b *testing.B) {
	b.StopTimer()

	r := NewRegistry[string, uint64, any]()

	err := r.Add(KeyTemporary, 1, "simple text insert benchmark")
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		var err error

		data, err = r.Get(KeyTemporary, 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEntity(b *testing.B) {
	b.StopTimer()

	r := NewRegistry[string, uint64, any]()

	err := r.Add(KeyTemporary, 1, &MockEntity{})
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		var err error

		data, err = r.Get(KeyTemporary, 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetString(b *testing.B) {
	r := NewRegistry[string, uint64, any]()
	for i := 0; i < b.N; i++ {
		id := r.NextID()

		err := r.Add(KeyTemporary, id, "simple text insert benchmark")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetEntity(b *testing.B) {
	r := NewRegistry[string, uint64, any]()
	for i := 0; i < b.N; i++ {
		id := r.NextID()

		err := r.Add(KeyTemporary, id, &MockEntity{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetDeleteString(b *testing.B) {
	r := NewRegistry[string, uint64, any]()
	for i := 0; i < b.N; i++ {
		id := r.NextID()

		err := r.Add(KeyTemporary, id, "simple text insert benchmark")
		if err != nil {
			b.Fatal(err)
		}

		err = r.Remove(KeyTemporary, id)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSetDeleteEntity(b *testing.B) {
	r := NewRegistry[string, uint64, any]()
	for i := 0; i < b.N; i++ {
		id := r.NextID()

		err := r.Add(KeyTemporary, id, &MockEntity{})
		if err != nil {
			b.Fatal(err)
		}

		err = r.Remove(KeyTemporary, id)
		if err != nil {
			b.Fatal(err)
		}
	}
}
