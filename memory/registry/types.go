package registry

// Nothing empty struct
var Nothing = struct{}{}

// Count of workers
const workersCount = 4

// Channel buffer size
const bufferSize = 1000

// Destroyable able to be destroyed in memory
type Destroyable interface {
	Destroy() error
}

// Constructable able to be construct in memory
type Constructable interface {
	Construct() error
}

// SearchFunction describe specification for filter function
type SearchFunction func(key interface{}, id interface{}, data interface{}) bool

// Ticker describe function for execute tick
type Ticker interface {
	Tick()
}
