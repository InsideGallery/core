package gremlin

import (
	"errors"
	"sync"

	"github.com/InsideGallery/core/memory/registry"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var ErrCastType = errors.New("error cast type")

type Cache struct {
	*registry.Registry[string, string, any]
	mu sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		Registry: registry.NewRegistry[string, string, any](),
	}
}

func (c *Cache) AddVertex(label, id string, vertex *gremlingo.Vertex) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Add(label, id, vertex)
}

func (c *Cache) GetVertex(label, id string) (*gremlingo.Vertex, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rawVertex, err := c.Registry.Get(label, id)
	if err != nil {
		return nil, err
	}

	vertex, ok := rawVertex.(*gremlingo.Vertex)
	if !ok {
		return nil, ErrCastType
	}

	return vertex, nil
}

func (c *Cache) DeleteVertex(label, id string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.Registry.Remove(label, id)

	return err
}

func (c *Cache) Truncate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Registry = registry.NewRegistry[string, string, any]()
}
