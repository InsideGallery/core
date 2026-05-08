package oslistener

import (
	"os"
	"sync"
)

// SignalListener contains signal and callbacks
type SignalListener struct {
	callbacks map[os.Signal][]func()
	mu        sync.Mutex
}

var (
	defaultListenerMu sync.RWMutex
	defaultListener   = NewSignalListener()
)

// NewSignalListener return new signal listener
func NewSignalListener() *SignalListener {
	return &SignalListener{
		callbacks: map[os.Signal][]func(){},
	}
}

// DefaultListener returns the package-level compatibility signal listener.
func DefaultListener() *SignalListener {
	defaultListenerMu.RLock()
	defer defaultListenerMu.RUnlock()

	return defaultListener
}

// DefaultListenerHandle restores a previous package-level signal listener.
type DefaultListenerHandle struct {
	previous *SignalListener
	once     sync.Once
}

// InstallDefaultListener installs a scoped package-level signal listener.
func InstallDefaultListener(listener *SignalListener) *DefaultListenerHandle {
	defaultListenerMu.Lock()
	defer defaultListenerMu.Unlock()

	if listener == nil {
		listener = NewSignalListener()
	}

	previous := defaultListener
	defaultListener = listener

	return &DefaultListenerHandle{
		previous: previous,
	}
}

// Close restores the previous package-level signal listener.
func (h *DefaultListenerHandle) Close() error {
	if h == nil {
		return nil
	}

	h.once.Do(func() {
		defaultListenerMu.Lock()

		defaultListener = h.previous

		defaultListenerMu.Unlock()
	})

	return nil
}

// Get returns the package-level compatibility signal listener.
//
// Deprecated: use NewSignalListener for explicit ownership or DefaultListener for compatibility.
func Get() *SignalListener {
	return DefaultListener()
}

// Append signal to listen
func (l *SignalListener) Append(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = append(l.callbacks[signal], fn)
}

// Prepend signal to listen
func (l *SignalListener) Prepend(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = append([]func(){fn}, l.callbacks[signal]...)
}

// Set signal to listen
func (l *SignalListener) Set(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = []func(){fn}
}

// Reset signal to listen
func (l *SignalListener) Reset(signal os.Signal) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = []func(){}
}

// Get signal to listen
func (l *SignalListener) Get(signal os.Signal) []func() {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.callbacks[signal]
}

// Wrap signal to listen
func (l *SignalListener) Wrap(signal os.Signal, fn func(...func()) func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = []func(){fn(l.callbacks[signal]...)}
}

// SignalsToSubscribe return list of signals
func (l *SignalListener) SignalsToSubscribe() OsSignalsList {
	l.mu.Lock()
	defer l.mu.Unlock()

	signals := make(OsSignalsList, len(l.callbacks))

	var i int

	for s := range l.callbacks {
		signals[i] = s
		i++
	}

	return signals
}

// ReceiveSignal call when signal received
func (l *SignalListener) ReceiveSignal(s os.Signal) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fns, ok := l.callbacks[s]
	if ok {
		for _, fn := range fns {
			fn()
		}
	}
}
