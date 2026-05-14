# ticker

Import path: `github.com/InsideGallery/core/ticker`

## Overview

`ticker` provides a global tick counter, a periodic handler manager, and a delayed one-shot executor.

## Main APIs

- `Tick`, `Get`, and `Reset` manage the package-level atomic tick counter.
- `Handler` is implemented by periodic workers with `Tick(context.Context)` and `GetID()`.
- `TickHandler` combines a context, interval, cancel function, and handler.
- `NewTickHandler` creates a cancellable handler wrapper.
- `TickManager` manages handlers with `Add`, `Remove`, `Stop`, `Handlers`, `GetHandlers`, `Run`, and
  `CountTicksInProgress`.
- `ExecuteWithDelay` runs a callback after a delay with `Start`, `Stop`, and `IsActive`.

## Usage

```go
type heartbeat struct {
	id uint64
}

func (h heartbeat) ID() uint64 { return h.id }
func (h heartbeat) GetID() uint64 { return h.id }
func (h heartbeat) Tick(context.Context) {}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

manager := ticker.NewTickManager()
manager.Add(ticker.NewTickHandler(ctx, time.Second, heartbeat{id: 1}))

go manager.Run()
```

## Notes

`Run` blocks until all handler loops finish. Each interval tick runs the handler in its own goroutine with a
context timeout equal to the handler interval. `Stop` cancels all handlers, and `Remove` cancels and deletes one
handler. `ExecuteWithDelay.Start` ignores a new start while the executor is already active.
