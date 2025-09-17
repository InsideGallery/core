package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/InsideGallery/core/memory/utils"
)

const waitTimeout = 10 * time.Millisecond

type Aggregator[K any] struct {
	mu  sync.RWMutex
	ctx context.Context

	ticker     time.Duration
	maxCount   int
	goroutines int
	cancel     context.CancelFunc
	closed     bool
	counter    chan struct{}

	items     *utils.SafeList[K]
	processor func([]K) error
}

func NewAggregator[K any](
	ctx context.Context, goroutines, count int, ticker time.Duration, processor func([]K) error,
) *Aggregator[K] {
	if goroutines <= 0 {
		goroutines = 1
	}

	if count <= 0 {
		goroutines = 1
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Aggregator[K]{
		ctx:        ctx,
		cancel:     cancel,
		maxCount:   count,
		goroutines: goroutines,
		ticker:     ticker,

		items:     utils.NewSafeList[K](),
		processor: processor,
		counter:   make(chan struct{}),
	}
}

func (w *Aggregator[K]) Add(req K) {
	w.mu.RLock() // This is special mutex, which should not block us on read
	defer w.mu.RUnlock()

	if w.closed {
		return
	}

	w.items.Add(req)

	if w.items.Count() >= w.maxCount {
		w.counter <- struct{}{}
	}
}

func (w *Aggregator[K]) Process() error {
	list := w.items.Reset()

	if len(list) == 0 || w.processor == nil {
		return nil
	}

	return w.processor(list)
}

func (w *Aggregator[K]) Count() int {
	return w.items.Count()
}

func (w *Aggregator[K]) Close() {
	w.cancel()
}

func (w *Aggregator[K]) Flusher() error {
	tck := time.NewTicker(w.ticker)

	var resultErr error

	RunSyncMultipleWorkers(w.ctx, w.goroutines, func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				w.mu.Lock()
				w.closed = true

				err := w.Process()
				if err != nil {
					slog.Default().Error("Error during flush by context", "err", err)
					resultErr = err

					w.mu.Unlock()

					return
				}

				w.mu.Unlock()

				return
			case <-tck.C:
				slog.Debug("Flush by ticker")

				err := w.Process()
				if err != nil {
					slog.Default().Error("Error during flush by ticker", "err", err)
				}
			case <-w.counter:
				slog.Debug("Flush by counter")

				err := w.Process()
				if err != nil {
					slog.Default().Error("Error during default flush by ticker", "err", err)
				}
			}
		}
	})

	return resultErr
}
