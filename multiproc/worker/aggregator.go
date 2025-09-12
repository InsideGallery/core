package worker

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/InsideGallery/core/memory/utils"
)

const (
	DefaultCounterCheck = 10 * time.Millisecond
)

func GetGoroutinesCount(variables, maxGoroutines int) int {
	res := slices.Min([]int{variables, maxGoroutines})
	if res <= 0 {
		return 1
	}

	return res
}

type Aggregator[K any] struct {
	mu  sync.RWMutex
	ctx context.Context

	ticker   time.Duration
	maxCount int
	cancel   context.CancelFunc
	closed   bool
	counter  chan struct{}

	items     *utils.SafeList[K]
	processor func([]K) error
}

func NewAggregator[K any](
	ctx context.Context, count int, ticker time.Duration, processor func([]K) error,
) *Aggregator[K] {
	ctx, cancel := context.WithCancel(ctx)

	return &Aggregator[K]{
		ctx:      ctx,
		cancel:   cancel,
		maxCount: count,
		ticker:   ticker,

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

	for {
		select {
		case <-w.ctx.Done():
			w.mu.Lock()
			w.closed = true
			err := w.Process()
			w.mu.Unlock()

			return err
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
}
