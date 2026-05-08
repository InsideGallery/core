package commands

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/InsideGallery/core/multiproc/worker"
)

const (
	minWorkersCount     = 1
	maxWorkersCount     = 5
	defaultWorkersCount = 4
)

// EventHandlerFunc event handler like function
type EventHandlerFunc func(ctx context.Context)

// Handle for event handler function
func (f EventHandlerFunc) Handle(ctx context.Context) {
	f(ctx)
}

// EventHandler describe event handler
type EventHandler interface {
	Handle(context.Context)
}

// EventManager event manager
type EventManager struct {
	subscribers map[string]map[uint64]EventHandler
	aid         uint64
	workers     int
	mu          sync.RWMutex
}

// NewEventManager return new event manager
func NewEventManager(workers int) *EventManager {
	if workers < minWorkersCount || workers > maxWorkersCount {
		workers = minWorkersCount
	}

	return &EventManager{
		subscribers: make(map[string]map[uint64]EventHandler),
		workers:     workers,
	}
}

// NextID return next id
func (e *EventManager) NextID() uint64 {
	return atomic.AddUint64(&e.aid, 1)
}

// Subscribe subscribe on event
func (e *EventManager) Subscribe(event string, handler EventHandler) uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.subscribers[event]; !ok {
		e.subscribers[event] = make(map[uint64]EventHandler)
	}

	id := e.NextID()
	e.subscribers[event][id] = handler

	return id
}

// Unsubscribe unsubscribe from event
func (e *EventManager) Unsubscribe(event string, id uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.subscribers[event]; ok {
		delete(e.subscribers[event], id)
	}

	if len(e.subscribers[event]) == 0 {
		delete(e.subscribers, event)
	}
}

// GetHandlers return all handlers for event
func (e *EventManager) GetHandlers(event string) []EventHandler {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, ok := e.subscribers[event]; !ok {
		return []EventHandler{}
	}

	var i int

	handlers := make([]EventHandler, len(e.subscribers[event]))

	for _, handler := range e.subscribers[event] {
		handlers[i] = handler
		i++
	}

	return handlers
}

// Call call event
func (e *EventManager) Call(ctx context.Context, event string) {
	defer func() {
		if rval := recover(); rval != nil {
			slog.Default().Error("Recovered request panic", "rval", rval)
		}
	}()

	handlers := e.GetHandlers(event)
	if len(handlers) == 0 {
		return
	}

	handler := make(chan EventHandler)

	go func() {
		for _, s := range handlers {
			handler <- s
		}

		close(handler)
	}()

	worker.RunSyncMultipleWorkers(ctx, e.workers, func(ctx context.Context) {
		for s := range handler {
			s.Handle(ctx)
		}
	})
}

var (
	defaultEventManagerMu sync.RWMutex
	defaultEventManager   = NewEventManager(defaultWorkersCount)
)

// DefaultEventManager returns the package-level compatibility event manager.
func DefaultEventManager() *EventManager {
	defaultEventManagerMu.RLock()
	defer defaultEventManagerMu.RUnlock()

	return defaultEventManager
}

// DefaultEventManagerHandle restores a previous package-level event manager.
type DefaultEventManagerHandle struct {
	previous *EventManager
	once     sync.Once
}

// InstallDefaultEventManager installs a scoped package-level event manager.
func InstallDefaultEventManager(manager *EventManager) *DefaultEventManagerHandle {
	defaultEventManagerMu.Lock()
	defer defaultEventManagerMu.Unlock()

	if manager == nil {
		manager = NewEventManager(defaultWorkersCount)
	}

	previous := defaultEventManager
	defaultEventManager = manager

	return &DefaultEventManagerHandle{
		previous: previous,
	}
}

// Close restores the previous package-level event manager.
func (h *DefaultEventManagerHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		defaultEventManagerMu.Lock()
		defaultEventManager = h.previous
		defaultEventManagerMu.Unlock()
	})

	return nil
}

// GetEventManager return default event manager
//
// Deprecated: use NewEventManager for explicit ownership or DefaultEventManager for compatibility.
func GetEventManager() *EventManager {
	return DefaultEventManager()
}
