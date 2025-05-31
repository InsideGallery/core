package registry

import (
	"sync"

	"github.com/InsideGallery/core/memory/orderedmap"
)

// Group contains all entities for given type
type Group[K comparable, V any] struct {
	entities *orderedmap.OrderedMap[K, V]
}

// NewGroup return registry
func NewGroup[K comparable, V any]() *Group[K, V] {
	return &Group[K, V]{
		entities: &orderedmap.OrderedMap[K, V]{},
	}
}

func (g *Group[K, V]) Add(id K, e V) (err error) {
	g.entities.Add(id, e)

	return g.construct(e)
}

func (g *Group[K, V]) construct(e any) (err error) {
	c, ok := e.(Constructable)
	if ok {
		err = c.Construct()
		if err != nil {
			return err
		}
	}

	return
}

func (g *Group[K, V]) Get(id K) (e V, err error) {
	if !g.entities.Exists(id) {
		err = ErrNotFoundEntity
		return
	}

	return g.entities.Get(id), nil
}

func (g *Group[K, V]) Size() int {
	return g.entities.Size()
}

func (g *Group[K, V]) CallWithLock(f func(d map[K]V) (e V, err error)) (V, error) {
	e, err := f(g.entities.GetMap())
	return e, err
}

func (g *Group[K, V]) Remove(id K) (err error) {
	exists := g.entities.Exists(id)
	e := g.entities.Get(id)
	g.entities.Remove(id)

	if exists {
		return g.destroy(e)
	}

	return
}

func (g *Group[K, V]) destroy(e any) (err error) {
	destroyable, ok := e.(Destroyable)
	if ok {
		err = destroyable.Destroy()
	}

	return
}

func (g *Group[K, V]) GetValues() []V {
	_, values := g.entities.GetAll()
	return values
}

func (g *Group[K, V]) GetKeys() []K {
	keys, _ := g.entities.GetAll()
	return keys
}

func (g *Group[K, V]) Iterator() chan V {
	return g.entities.Iterator(bufferSize)
}

func (g *Group[K, V]) Truncate() {
	g.entities.Truncate()
}

func (g *Group[K, V]) GetMap() map[K]V {
	return g.entities.GetMap()
}

func (g *Group[K, V]) Tick() {
	ch := g.entities.Iterator(bufferSize)
	for entity := range ch {
		g.tick(entity)
	}
}

func (g *Group[K, V]) tick(v any) {
	n, ok := v.(Ticker)
	if ok {
		n.Tick()
	}
}

type item[K comparable, V any] struct {
	id   K
	data V
}

func (g *Group[K, V]) Search(key any, result chan V, f SearchFunction) {
	entities := make(chan item[K, V], bufferSize)

	var wg sync.WaitGroup
	wg.Add(workersCount)

	for i := 0; i < workersCount; i++ {
		go func(result chan V, f SearchFunction) {
			for n := range entities {
				if f == nil || f(key, n.id, n.data) {
					result <- n.data
				}
			}

			wg.Done()
		}(result, f)
	}

	for id, entity := range g.GetMap() {
		entities <- item[K, V]{
			id:   id,
			data: entity,
		}
	}

	close(entities)
	wg.Wait()
}

func (g *Group[K, V]) SearchOne(key any, f SearchFunction) (d V) {
	for id, data := range g.GetMap() {
		if f == nil || f(key, id, data) {
			d = data
			break
		}
	}

	return
}
