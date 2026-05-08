package oslistener

import (
	"reflect"
	"syscall"
	"testing"
)

func TestSignalListenerOperations(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "append prepend and receive signal",
			run: func(t *testing.T) {
				t.Helper()

				listener := NewSignalListener()
				var calls []string

				listener.Append(syscall.SIGUSR1, func() {
					calls = append(calls, "second")
				})
				listener.Prepend(syscall.SIGUSR1, func() {
					calls = append(calls, "first")
				})
				listener.ReceiveSignal(syscall.SIGUSR1)

				if !reflect.DeepEqual(calls, []string{"first", "second"}) {
					t.Fatalf("calls = %v", calls)
				}
			},
		},
		{
			name: "set reset and get callbacks",
			run: func(t *testing.T) {
				t.Helper()

				listener := NewSignalListener()
				listener.Append(syscall.SIGUSR1, func() {})
				listener.Set(syscall.SIGUSR1, func() {})

				if got := len(listener.Get(syscall.SIGUSR1)); got != 1 {
					t.Fatalf("callbacks = %d, want 1", got)
				}

				listener.Reset(syscall.SIGUSR1)
				if got := len(listener.Get(syscall.SIGUSR1)); got != 0 {
					t.Fatalf("callbacks after reset = %d, want 0", got)
				}
			},
		},
		{
			name: "wrap callbacks",
			run: func(t *testing.T) {
				t.Helper()

				listener := NewSignalListener()
				var calls []string
				listener.Append(syscall.SIGUSR1, func() {
					calls = append(calls, "inner")
				})
				listener.Wrap(syscall.SIGUSR1, func(fns ...func()) func() {
					return func() {
						calls = append(calls, "before")
						for _, fn := range fns {
							fn()
						}
					}
				})

				listener.ReceiveSignal(syscall.SIGUSR1)
				if !reflect.DeepEqual(calls, []string{"before", "inner"}) {
					t.Fatalf("calls = %v", calls)
				}
			},
		},
		{
			name: "signals to subscribe returns registered signals",
			run: func(t *testing.T) {
				t.Helper()

				listener := NewSignalListener()
				listener.Append(syscall.SIGUSR1, func() {})
				listener.Append(syscall.SIGUSR2, func() {})

				signals := listener.SignalsToSubscribe()
				if len(signals) != 2 {
					t.Fatalf("signals = %v, want 2", signals)
				}
			},
		},
		{
			name: "default listener handle restores previous listener",
			run: func(t *testing.T) {
				t.Helper()

				previous := DefaultListener()
				next := NewSignalListener()
				handle := InstallDefaultListener(next)

				if got := Get(); got != next {
					t.Fatal("default listener was not installed")
				}

				if err := handle.Close(); err != nil {
					t.Fatalf("close default listener handle: %v", err)
				}

				if got := DefaultListener(); got != previous {
					t.Fatal("default listener was not restored")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
