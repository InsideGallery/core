package oslistener

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

type TestListener struct {
	// Dont use a map, since its not thread safe
	MessageReceived chan os.Signal
}

func (testListener *TestListener) SignalsToSubscribe() OsSignalsList {
	return OsSignalsList{syscall.SIGHUP, syscall.SIGINT}
}

func (testListener *TestListener) ReceiveSignal(signal os.Signal) {
	testListener.MessageReceived <- signal
}

func TestStart(t *testing.T) {
	received := make(chan os.Signal)
	signalsReceived := map[os.Signal]bool{
		syscall.SIGHUP: true,
		syscall.SIGINT: true,
		// This one should not be received
		syscall.SIGUSR1: true,
	}
	tl := &TestListener{MessageReceived: received}

	// Max execution time 1 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	Start(ctx, tl)

	err := Raise(syscall.SIGHUP)
	if err != nil {
		t.Fatal(err)
	}

	err = Raise(syscall.SIGUSR1)
	if err != nil {
		t.Fatal(err)
	}

	err = Raise(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}

	for s := range tl.MessageReceived {
		delete(signalsReceived, s)

		if len(signalsReceived) == 1 {
			break
		}
	}

	if len(signalsReceived) != 1 {
		t.Fatalf("Didnt receive some signals: %v", signalsReceived)
	}
}
