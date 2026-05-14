package oslistener

import (
	"os"
	"syscall"
	"testing"
)

func TestNewSignalListener(t *testing.T) {
	sl := NewSignalListener()
	if sl == nil {
		t.Fatal("expected non-nil SignalListener")
	}

	// Should start with no signals to subscribe.
	sigs := sl.SignalsToSubscribe()
	if len(sigs) != 0 {
		t.Errorf("expected 0 signals, got %d", len(sigs))
	}
}

func TestGet(t *testing.T) {
	sl := Get()
	if sl == nil {
		t.Fatal("expected non-nil default SignalListener")
	}
}

func TestAppendAndReceiveSignal(t *testing.T) {
	sl := NewSignalListener()

	called := false

	sl.Append(syscall.SIGUSR1, func() {
		called = true
	})

	sigs := sl.SignalsToSubscribe()
	if len(sigs) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(sigs))
	}

	sl.ReceiveSignal(syscall.SIGUSR1)

	if !called {
		t.Error("expected callback to be called")
	}
}

func TestAppendMultipleCallbacks(t *testing.T) {
	sl := NewSignalListener()

	count := 0

	sl.Append(syscall.SIGUSR1, func() { count++ })
	sl.Append(syscall.SIGUSR1, func() { count++ })

	sl.ReceiveSignal(syscall.SIGUSR1)

	if count != 2 {
		t.Errorf("expected 2 callbacks, got %d", count)
	}
}

func TestPrepend(t *testing.T) {
	sl := NewSignalListener()

	var order []int

	sl.Append(syscall.SIGUSR1, func() { order = append(order, 2) })
	sl.Prepend(syscall.SIGUSR1, func() { order = append(order, 1) })

	sl.ReceiveSignal(syscall.SIGUSR1)

	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("expected [1, 2], got %v", order)
	}
}

func TestSet(t *testing.T) {
	sl := NewSignalListener()

	count := 0

	sl.Append(syscall.SIGUSR1, func() { count++ })
	sl.Append(syscall.SIGUSR1, func() { count++ })

	// Set replaces all callbacks with a single one.
	sl.Set(syscall.SIGUSR1, func() { count = 42 })

	sl.ReceiveSignal(syscall.SIGUSR1)

	if count != 42 {
		t.Errorf("expected 42, got %d", count)
	}
}

func TestReset(t *testing.T) {
	sl := NewSignalListener()

	called := false

	sl.Append(syscall.SIGUSR1, func() { called = true })

	sl.Reset(syscall.SIGUSR1)
	sl.ReceiveSignal(syscall.SIGUSR1)

	if called {
		t.Error("expected callback NOT to be called after Reset")
	}
}

func TestReceiveSignalNoCallbacks(_ *testing.T) {
	sl := NewSignalListener()

	// Should not panic.
	sl.ReceiveSignal(syscall.SIGUSR1)
}

func TestReceiveSignalDifferentSignal(t *testing.T) {
	sl := NewSignalListener()

	called := false

	sl.Append(syscall.SIGUSR1, func() { called = true })

	// Receive a different signal.
	sl.ReceiveSignal(syscall.SIGUSR2)

	if called {
		t.Error("callback should not be called for a different signal")
	}
}

func TestMultipleSignals(t *testing.T) {
	sl := NewSignalListener()

	var sig1Called, sig2Called bool

	sl.Append(syscall.SIGUSR1, func() { sig1Called = true })
	sl.Append(syscall.SIGUSR2, func() { sig2Called = true })

	sigs := sl.SignalsToSubscribe()
	if len(sigs) != 2 {
		t.Errorf("expected 2 signals, got %d", len(sigs))
	}

	sl.ReceiveSignal(syscall.SIGUSR1)

	if !sig1Called {
		t.Error("expected SIGUSR1 callback to be called")
	}

	if sig2Called {
		t.Error("expected SIGUSR2 callback NOT to be called")
	}
}

func TestSignalsToSubscribeReturnsRegisteredOnly(t *testing.T) {
	sl := NewSignalListener()
	sl.Append(syscall.SIGINT, func() {})
	sl.Append(syscall.SIGTERM, func() {})

	sigs := sl.SignalsToSubscribe()
	if len(sigs) != 2 {
		t.Errorf("expected 2 signals, got %d", len(sigs))
	}

	sigMap := make(map[os.Signal]bool)
	for _, s := range sigs {
		sigMap[s] = true
	}

	if !sigMap[syscall.SIGINT] {
		t.Error("expected SIGINT in signals list")
	}

	if !sigMap[syscall.SIGTERM] {
		t.Error("expected SIGTERM in signals list")
	}
}
