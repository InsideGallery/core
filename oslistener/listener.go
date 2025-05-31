package oslistener

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
)

// OsSignalsList is a collection of signals
type OsSignalsList []os.Signal

// OsListener is an interface which allows the object to listen to certain signals
type OsListener interface {
	SignalsToSubscribe() OsSignalsList
	ReceiveSignal(os.Signal)
}

// Start will start a Go routine in the background, which listens to OS signals
func Start(ctx context.Context, listener OsListener) {
	signalsForSubscription := listener.SignalsToSubscribe()
	sigs := make(chan os.Signal, len(signalsForSubscription))
	signal.Notify(sigs, signalsForSubscription...)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Default().Warn("Stopping the signal listener")
				return
			case receivedSignal := <-sigs:
				listener.ReceiveSignal(receivedSignal)
			}
		}
	}()
}

// Raise sends the given signal to the current process
func Raise(sig os.Signal) error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}

	return p.Signal(sig)
}
