package ecs

import (
	"sync/atomic"

	"github.com/InsideGallery/core/memory/registry"
)

var store = registry.NewRegistry[string, string, any]()

// BaseEntity contains required fields
type BaseEntity struct {
	id      uint64
	version uint64
}

// NewBaseEntity return new entity
func NewBaseEntity() *BaseEntity {
	return &BaseEntity{
		id:      store.NextID(),
		version: 1,
	}
}

// NewBaseEntityWithID return new entity with id
func NewBaseEntityWithID(id uint64) *BaseEntity {
	return &BaseEntity{
		id:      id,
		version: 1,
	}
}

// GetID return id
func (e *BaseEntity) GetID() uint64 {
	return e.id
}

// SetID set id
func (e *BaseEntity) SetID(id uint64) {
	lid := store.LatestID()
	if id > lid {
		store.SetLatestID(id)
	}

	e.id = id
}

// GetVersion return current version of object
func (e *BaseEntity) GetVersion() uint64 {
	return atomic.LoadUint64(&e.version)
}

// UpVersion increase current version
func (e *BaseEntity) UpVersion() {
	atomic.AddUint64(&e.version, 1)
}

// SetVersion set current version
func (e *BaseEntity) SetVersion(v uint64) {
	atomic.StoreUint64(&e.version, v)
}
