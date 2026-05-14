package ecs

import (
	"sync"
	"sync/atomic"

	"github.com/FrogoAI/memory/registry"
)

// Registry owns ECS entity ID generation state for application composition.
type Registry struct {
	store *registry.Registry[string, string, any]
}

// EntityFactory aliases Registry for backward compatibility.
//
// Deprecated: use Registry.
type EntityFactory = Registry

var defaultRegistryMu sync.RWMutex

// Default is the package-level compatibility ECS registry used by legacy helpers.
var Default = NewRegistry()

// BaseEntity contains required fields
type BaseEntity struct {
	id        uint64
	version   uint64
	idFactory *EntityFactory
}

// NewRegistry returns an isolated ECS registry.
func NewRegistry() *Registry {
	return &Registry{
		store: registry.NewRegistry[string, string, any](),
	}
}

// NewEntityFactory returns an isolated ECS entity factory.
//
// Deprecated: use NewRegistry.
func NewEntityFactory() *EntityFactory {
	return NewRegistry()
}

// DefaultEntityFactory returns the package-level compatibility entity factory.
func DefaultEntityFactory() *EntityFactory {
	return defaultRegistry()
}

// DefaultEntityFactoryHandle restores a previous package-level entity factory.
type DefaultEntityFactoryHandle struct {
	previous *EntityFactory
	once     sync.Once
}

// InstallDefaultEntityFactory installs a scoped package-level entity factory.
func InstallDefaultEntityFactory(factory *EntityFactory) *DefaultEntityFactoryHandle {
	defaultRegistryMu.Lock()
	defer defaultRegistryMu.Unlock()

	if factory == nil || factory.store == nil {
		factory = NewRegistry()
	}

	previous := Default
	if previous == nil || previous.store == nil {
		previous = NewRegistry()
	}

	Default = factory

	return &DefaultEntityFactoryHandle{
		previous: previous,
	}
}

// Close restores the previous package-level entity factory.
func (h *DefaultEntityFactoryHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		defaultRegistryMu.Lock()

		Default = h.previous

		defaultRegistryMu.Unlock()
	})

	return nil
}

// NewBaseEntity returns a new entity owned by this registry.
func (r *Registry) NewBaseEntity() *BaseEntity {
	ecsRegistry := validRegistry(r)

	return &BaseEntity{
		id:        ecsRegistry.nextID(),
		version:   1,
		idFactory: ecsRegistry,
	}
}

// NewBaseEntityWithID returns a new entity with an explicit ID owned by this registry.
func (r *Registry) NewBaseEntityWithID(id uint64) *BaseEntity {
	return &BaseEntity{
		id:        id,
		version:   1,
		idFactory: validRegistry(r),
	}
}

// LatestID returns the latest generated entity ID for this registry.
func (r *Registry) LatestID() uint64 {
	return validRegistry(r).store.LatestID()
}

// NewBaseEntity return new entity.
//
// Deprecated: use NewEntityFactory and EntityFactory.NewBaseEntity for explicit ID ownership.
func NewBaseEntity() *BaseEntity {
	return DefaultEntityFactory().NewBaseEntity()
}

// NewBaseEntityWithID return new entity with id.
//
// Deprecated: use NewEntityFactory and EntityFactory.NewBaseEntityWithID for explicit ID ownership.
func NewBaseEntityWithID(id uint64) *BaseEntity {
	return DefaultEntityFactory().NewBaseEntityWithID(id)
}

// GetID return id
func (e *BaseEntity) GetID() uint64 {
	return e.id
}

// SetID set id
func (e *BaseEntity) SetID(id uint64) {
	e.entityFactory().rememberID(id)

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

func defaultRegistry() *Registry {
	current := currentDefaultRegistry()
	if current != nil && current.store != nil {
		return current
	}

	defaultRegistryMu.Lock()
	defer defaultRegistryMu.Unlock()

	if Default == nil || Default.store == nil {
		Default = NewRegistry()
	}

	return Default
}

func currentDefaultRegistry() *Registry {
	defaultRegistryMu.RLock()
	defer defaultRegistryMu.RUnlock()

	return Default
}

func validRegistry(ecsRegistry *Registry) *Registry {
	if ecsRegistry == nil || ecsRegistry.store == nil {
		return defaultRegistry()
	}

	return ecsRegistry
}

func (e *BaseEntity) entityFactory() *Registry {
	if e.idFactory == nil || e.idFactory.store == nil {
		e.idFactory = defaultRegistry()
	}

	return e.idFactory
}

func (r *Registry) nextID() uint64 {
	return r.store.NextID()
}

func (r *Registry) rememberID(id uint64) {
	latestID := r.store.LatestID()
	if id > latestID {
		r.store.SetLatestID(id)
	}
}
