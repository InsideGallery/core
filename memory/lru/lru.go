package lru

import (
	"container/list"
	"sync"

	"github.com/spf13/cast"
)

type Value[K any] struct {
	Value   K
	Element *list.Element
}

type Cache[K any] struct {
	list     *list.List
	index    map[string]*Value[K]
	mu       sync.Mutex
	capacity int
}

func NewLRUCache[K any](capacity int) *Cache[K] {
	return &Cache[K]{
		list:     list.New(),
		index:    map[string]*Value[K]{},
		capacity: capacity,
	}
}

func (c *Cache[K]) Get(key string) (K, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var val K

	v, ok := c.index[key]
	if !ok {
		return val, false
	}

	c.list.MoveToFront(v.Element)

	return v.Value, true
}

func (c *Cache[K]) Put(key string, value K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.index[key]
	if !ok {
		c.list.PushFront(key)
	} else {
		c.list.MoveToFront(v.Element)
	}

	c.index[key] = &Value[K]{
		Value:   value,
		Element: c.list.Front(),
	}

	listSize := c.list.Len()

	if listSize >= c.capacity {
		prevKey := c.list.Back()
		c.list.Remove(prevKey)

		delete(c.index, cast.ToString(prevKey.Value))
	}
}
