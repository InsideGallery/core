# oslistener

Import path: `github.com/InsideGallery/core/oslistener`

## Overview

`oslistener` maps OS signals to callbacks and starts a goroutine that dispatches process signals to an
`OsListener`.

## Main APIs

- `OsSignalsList` is a slice of `os.Signal`.
- `OsListener` describes a signal source with `SignalsToSubscribe` and `ReceiveSignal`.
- `Start(ctx, listener)` subscribes with `signal.Notify` and dispatches received signals until the context ends.
- `Raise(sig)` sends a signal to the current process.
- `SignalListener` stores callbacks per signal.
- `NewSignalListener`, `Append`, `Prepend`, `Set`, `Reset`, `SignalsToSubscribe`, and `ReceiveSignal` manage
  callbacks.
- `Get`, `DefaultListener`, and `InstallDefaultListener` expose the package-level compatibility listener.

## Usage

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

listener := oslistener.NewSignalListener()
listener.Append(syscall.SIGTERM, cancel)

oslistener.Start(ctx, listener)
```

## Notes

`Start` returns immediately after launching its goroutine. Cancel the context to call `signal.Stop` for the
subscription channel. `SignalListener` executes registered callbacks synchronously while holding its internal
mutex, so callbacks should return promptly.
