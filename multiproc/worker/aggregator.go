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

type Aggregator struct {
	mu  sync.RWMutex
	ctx context.Context

	ticker   time.Duration
	maxCount int
	cancel   context.CancelFunc
	closed   bool

	items     *utils.SafeList[any]
	processor func([]any) error
}

func NewAggregator(ctx context.Context, count int, ticker time.Duration, processor func([]any) error) *Aggregator {
	ctx, cancel := context.WithCancel(ctx)

	return &Aggregator{
		ctx:      ctx,
		cancel:   cancel,
		maxCount: count,
		ticker:   ticker,

		items:     utils.NewSafeList[any](),
		processor: processor,
	}
}

func (w *Aggregator) Add(req any) {
	w.mu.RLock() // This is special mutex, which should not block us on read
	defer w.mu.RUnlock()

	if w.closed {
		return
	}

	w.items.Add(req)
}

func (w *Aggregator) Process() error {
	list := w.items.Reset()

	if len(list) == 0 || w.processor == nil {
		return nil
	}

	return w.processor(list)
}

func (w *Aggregator) Count() int {
	return w.items.Count()
}

func (w *Aggregator) Close() {
	w.cancel()
}

func (w *Aggregator) Flusher() error {
	tck := time.NewTicker(w.ticker)
	counterCheck := time.NewTicker(DefaultCounterCheck)

	for {
		select {
		case <-w.ctx.Done():
			w.mu.Lock()
			w.closed = true
			err := w.Process()
			w.mu.Unlock()

			return err
		case <-tck.C:
			err := w.Process()
			if err != nil {
				slog.Default().Error("Error during flush by ticker", "err", err)
			}
		case <-counterCheck.C:
			count := w.items.Count()
			if count > w.maxCount {
				err := w.Process()
				if err != nil {
					slog.Default().Error("Error during default flush by ticker", "err", err)
				}
			}
		}
	}
}
