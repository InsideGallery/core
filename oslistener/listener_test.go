package oslistener

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

// mockListener implements OsListener for testing.
type mockListener struct {
	mu       sync.Mutex
	signals  OsSignalsList
	received []os.Signal
}

func newMockListener(sigs ...os.Signal) *mockListener {
	return &mockListener{signals: sigs}
}

func (m *mockListener) SignalsToSubscribe() OsSignalsList {
	return m.signals
}

func (m *mockListener) ReceiveSignal(sig os.Signal) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.received = append(m.received, sig)
}

func (m *mockListener) receivedSignals() []os.Signal {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make([]os.Signal, len(m.received))
	copy(cp, m.received)

	return cp
}

func TestOsSignalsList(t *testing.T) {
	tests := []struct {
		name    string
		signals OsSignalsList
		wantLen int
	}{
		{
			name:    "empty list",
			signals: OsSignalsList{},
			wantLen: 0,
		},
		{
			name:    "single signal",
			signals: OsSignalsList{syscall.SIGUSR1},
			wantLen: 1,
		},
		{
			name:    "multiple signals",
			signals: OsSignalsList{syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(tt.signals); got != tt.wantLen {
				t.Errorf("len(OsSignalsList) = %d, want %d", got, tt.wantLen)
			}
		})
	}
}

func TestMockListenerSignalsToSubscribe(t *testing.T) {
	tests := []struct {
		name    string
		signals []os.Signal
		wantLen int
	}{
		{
			name:    "no signals",
			signals: nil,
			wantLen: 0,
		},
		{
			name:    "one signal",
			signals: []os.Signal{syscall.SIGUSR1},
			wantLen: 1,
		},
		{
			name:    "two signals",
			signals: []os.Signal{syscall.SIGUSR1, syscall.SIGUSR2},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := newMockListener(tt.signals...)
			got := ml.SignalsToSubscribe()

			if len(got) != tt.wantLen {
				t.Errorf("SignalsToSubscribe() len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestStartDispatchesSignal(t *testing.T) {
	ml := newMockListener(syscall.SIGUSR1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	Start(ctx, ml)

	// Give the goroutine time to register the signal handler.
	time.Sleep(50 * time.Millisecond)

	if err := Raise(syscall.SIGUSR1); err != nil {
		t.Fatalf("Raise(SIGUSR1): %v", err)
	}

	// Wait for the signal to be dispatched.
	deadline := time.After(2 * time.Second)

	for {
		if sigs := ml.receivedSignals(); len(sigs) > 0 {
			if sigs[0] != syscall.SIGUSR1 {
				t.Errorf("received signal = %v, want SIGUSR1", sigs[0])
			}

			return
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for signal dispatch")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestStartStopsOnContextCancel(t *testing.T) {
	ml := newMockListener(syscall.SIGUSR2)
	ctx, cancel := context.WithCancel(context.Background())

	Start(ctx, ml)

	// Give goroutine time to start.
	time.Sleep(50 * time.Millisecond)

	cancel()

	// Give goroutine time to stop.
	time.Sleep(50 * time.Millisecond)

	// After cancel, sending a signal should not be received by the listener.
	_ = Raise(syscall.SIGUSR2)

	time.Sleep(100 * time.Millisecond)

	if sigs := ml.receivedSignals(); len(sigs) > 0 {
		t.Errorf("expected no signals after cancel, got %d", len(sigs))
	}
}

func TestRaise(t *testing.T) {
	tests := []struct {
		name string
		sig  os.Signal
	}{
		{
			name: "raise SIGUSR1",
			sig:  syscall.SIGUSR1,
		},
		{
			name: "raise SIGUSR2",
			sig:  syscall.SIGUSR2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := newMockListener(tt.sig)
			ctx, cancel := context.WithCancel(context.Background())

			defer cancel()

			Start(ctx, ml)
			time.Sleep(50 * time.Millisecond)

			if err := Raise(tt.sig); err != nil {
				t.Fatalf("Raise(%v): %v", tt.sig, err)
			}

			deadline := time.After(2 * time.Second)

			for {
				if sigs := ml.receivedSignals(); len(sigs) > 0 {
					if sigs[0] != tt.sig {
						t.Errorf("received signal = %v, want %v", sigs[0], tt.sig)
					}

					return
				}

				select {
				case <-deadline:
					t.Fatal("timed out waiting for signal")
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
		})
	}
}
