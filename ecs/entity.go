package ecs

// Entity describe entity
type Entity interface {
	GetID() uint64
}

// Versionable describe version for entity
type Versionable interface {
	GetVersion() uint64
	UpVersion()
}
