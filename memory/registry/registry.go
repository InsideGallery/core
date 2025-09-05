package registry

import (
	"sync"
	"sync/atomic"

	"github.com/InsideGallery/core/errors"
)

const defaultBuffer = 1000

type Registry[G comparable, I comparable, V any] struct {
	groups  map[G]*Group[I, V]
	indexes map[uint64]I
	aid     uint64
	mu      sync.RWMutex
}

func NewRegistry[G comparable, I comparable, V any]() *Registry[G, I, V] {
	return &Registry[G, I, V]{
		groups:  make(map[G]*Group[I, V]),
		indexes: make(map[uint64]I),
	}
}

func (r *Registry[G, I, V]) NextID() uint64 {
	return atomic.AddUint64(&r.aid, 1)
}

func (r *Registry[G, I, V]) LatestID() uint64 {
	return atomic.LoadUint64(&r.aid)
}

func (r *Registry[G, I, V]) SetLatestID(id uint64) {
	atomic.StoreUint64(&r.aid, id)
}

func (r *Registry[G, I, V]) GetGroup(key G) (group *Group[I, V]) {
	var exists bool

	r.mu.RLock()
	group, exists = r.groups[key]
	r.mu.RUnlock()

	if !exists {
		group = r.initGroup(key)
	}

	return
}

func (r *Registry[G, I, V]) GetGroups(keys ...G) (groups []*Group[I, V]) {
	for _, key := range keys {
		groups = append(groups, r.GetGroup(key))
	}

	return
}

func (r *Registry[G, I, V]) AsyncIterator(keys ...G) chan V {
	result := make(chan V, bufferSize)

	var wg sync.WaitGroup
	wg.Add(len(keys))

	for _, key := range keys {
		go func(key G) {
			defer wg.Done()

			ch := r.GetGroup(key).Iterator()

			for item := range ch {
				result <- item
			}
		}(key)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func (r *Registry[G, I, V]) Iterator(keys ...G) chan V {
	result := make(chan V, bufferSize)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for _, key := range keys {
			ch := r.GetGroup(key).Iterator()
			for item := range ch {
				result <- item
			}
		}
	}()

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func (r *Registry[G, I, V]) initGroup(key G) (group *Group[I, V]) {
	var exists bool

	r.mu.Lock()
	defer r.mu.Unlock()

	group, exists = r.groups[key]

	if !exists {
		group = NewGroup[I, V]()
		r.groups[key] = group
	}

	return
}

func (r *Registry[G, I, V]) DeleteGroup(key G) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.groups, key)
}

func (r *Registry[G, I, V]) AddIndex(id uint64, key I) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.indexes[id] = key
}

func (r *Registry[G, I, V]) GetIndex(id uint64) I {
	r.mu.RLock()
	defer r.mu.RUnlock()

	i := r.indexes[id]

	return i
}

func (r *Registry[G, I, V]) RemIndex(id uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.indexes, id)
}

func (r *Registry[G, I, V]) Add(key G, id I, e V) error {
	group := r.GetGroup(key)
	return group.Add(id, e)
}

func (r *Registry[G, I, V]) Get(key G, id I) (e V, err error) {
	group := r.GetGroup(key)
	return group.Get(id)
}

func (r *Registry[G, I, V]) Remove(key G, id I) error {
	group := r.GetGroup(key)
	return group.Remove(id)
}

func (r *Registry[G, I, V]) RemoveIDEverywhere(id I) error {
	var errs []error

	groups := r.GetGroups(r.GetKeys()...)

	for _, group := range groups {
		errs = append(errs, group.Remove(id))
	}

	return errors.Combine(errs...)
}

func (r *Registry[G, I, V]) GetValues(key G) []V {
	group := r.GetGroup(key)
	return group.GetValues()
}

func (r *Registry[G, I, V]) TickGroup(key G) {
	group := r.GetGroup(key)
	group.Tick()
}

func (r *Registry[G, I, V]) TickGroups(keys ...G) {
	for _, key := range keys {
		r.TickGroup(key)
	}
}

func (r *Registry[G, I, V]) AsyncTick(keys ...G) {
	var wg sync.WaitGroup
	wg.Add(len(keys))

	for _, key := range keys {
		go func(key G) {
			r.TickGroup(key)
			wg.Done()
		}(key)
	}

	wg.Wait()
}

func (r *Registry[G, I, V]) TruncateGroup(key G) {
	group := r.GetGroup(key)
	group.Truncate()
}

func (r *Registry[G, I, V]) SearchInGroup(key G, f SearchFunction) chan V {
	result := make(chan V, defaultBuffer)
	group := r.GetGroup(key)

	go func(key G, result chan V, f SearchFunction) {
		group.Search(key, result, f)
		close(result)
	}(key, result, f)

	return result
}

func (r *Registry[G, I, V]) SearchOne(key G, f SearchFunction) V {
	group := r.GetGroup(key)
	return group.SearchOne(key, f)
}

func (r *Registry[G, I, V]) GetKeys() []G {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]G, 0, len(r.groups))
	for k := range r.groups {
		res = append(res, k)
	}

	return res
}

func (r *Registry[G, I, V]) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.groups)
}
