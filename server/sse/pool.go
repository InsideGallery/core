package sse

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/InsideGallery/core/memory/set"
	"github.com/InsideGallery/core/utils"
)

const defaultBufferSize = 1000

// Describe all context keys
var (
	ContextUserID utils.ContextKey = "userID"
)

// Pool contains all connections and could delivery messages
type Pool struct {
	mu          *sync.RWMutex
	connections map[string]chan Message
	bufferSize  int
}

// NewPool return new handler
func NewPool(bufferSize int) *Pool {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}

	return &Pool{
		connections: map[string]chan Message{},
		mu:          &sync.RWMutex{},
		bufferSize:  bufferSize,
	}
}

// Connections return count of connections
func (h *Pool) Connections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.connections)
}

// Add add new connection
func (h *Pool) Add(userID string) chan Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Message, h.bufferSize)
	h.connections[userID] = ch

	return ch
}

// Remove remove connection
func (h *Pool) Remove(userID string) {
	h.mu.Lock()
	ch := h.connections[userID]
	delete(h.connections, userID)
	close(ch)
	h.mu.Unlock()
}

// GetAllConnectedUsers get all connected users
func (h *Pool) GetAllConnectedUsers() set.GenericDataSet[string] {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connections := set.NewGenericDataSet[string]()
	for key := range h.connections {
		connections.Add(key)
	}

	return connections
}

// StopAll stop all connections
func (h *Pool) StopAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for key := range h.connections {
		ch := h.connections[key]
		delete(h.connections, key)
		close(ch)
	}
}

// SendToAll send message to all connections
func (h *Pool) SendToAll(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.connections {
		c <- msg
	}
}

// Send send message to all connections
func (h *Pool) Send(userID string, msg Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ch, ok := h.connections[userID]
	if !ok {
		return ErrNotFoundConnectedUser
	}

	ch <- msg

	return nil
}

// Handler listen for messages
func (h *Pool) Handler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextUserID).(string)
	if !ok {
		slog.Default().Error("Error getting user id from context", "err", ErrInvalidUserID)
		return
	}

	ch := h.Add(userID)

	defer func() {
		h.Remove(userID)
	}()

	err := Run(ch, w, r)
	if err != nil {
		slog.Default().Error("Error processing sse", "err", err)
		return
	}
}
