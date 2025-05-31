package registry

import (
	"strings"
	"sync"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

const KeyTemporary = "test"

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
