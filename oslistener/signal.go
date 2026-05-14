package oslistener

import (
	"os"
	"sync"
)

var (
	defaultListenerMu sync.RWMutex
	defaultListener   = NewSignalListener()
)

// Get returns the package-level default SignalListener.
func Get() *SignalListener {
	return DefaultListener()
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

// SignalListener maps OS signals to callback functions.
type SignalListener struct {
	callbacks map[os.Signal][]func()
	mu        sync.Mutex
}

// NewSignalListener creates a new empty SignalListener.
func NewSignalListener() *SignalListener {
	return &SignalListener{
		callbacks: map[os.Signal][]func(){},
	}
}

// Append adds a callback to be executed when the signal is received.
func (l *SignalListener) Append(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = append(l.callbacks[signal], fn)
}

// Prepend adds a callback to be executed first when the signal is received.
func (l *SignalListener) Prepend(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = append([]func(){fn}, l.callbacks[signal]...)
}

// Set replaces all callbacks for the given signal with a single callback.
func (l *SignalListener) Set(signal os.Signal, fn func()) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = []func(){fn}
}

// Reset removes all callbacks for the given signal.
func (l *SignalListener) Reset(signal os.Signal) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.callbacks[signal] = []func(){}
}

// SignalsToSubscribe returns the list of signals that have registered callbacks.
func (l *SignalListener) SignalsToSubscribe() OsSignalsList {
	l.mu.Lock()
	defer l.mu.Unlock()

	signals := make(OsSignalsList, 0, len(l.callbacks))
	for s := range l.callbacks {
		signals = append(signals, s)
	}

	return signals
}

// ReceiveSignal executes all callbacks registered for the given signal.
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
