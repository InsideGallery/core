package ticker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

const maxDelayedTicks = 100

// Handler describe tick handler
type Handler interface {
	Tick(ctx context.Context)
	GetID() uint64
}

// TickHandler contains handler and interval
type TickHandler struct {
	ctx      context.Context
	handler  Handler
	cancel   func()
	interval time.Duration
}

// NewTickHandler return new tick handler
func NewTickHandler(ctx context.Context, interval time.Duration, handler Handler) *TickHandler {
	ctx, cancel := context.WithCancel(ctx)

	return &TickHandler{
		interval: interval,
		handler:  handler,
		cancel:   cancel,
		ctx:      ctx,
	}
}

// TickManager describe TickManager
type TickManager struct {
	ticker map[uint64]*TickHandler
	mu     *sync.RWMutex
	count  int32
}

// NewTickManager return tick manager
func NewTickManager() *TickManager {
	return &TickManager{
		ticker: map[uint64]*TickHandler{},
		mu:     &sync.RWMutex{},
	}
}

// Add add tick handler
func (t *TickManager) Add(h *TickHandler) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := h.handler.GetID()
	t.ticker[id] = h

	return id
}

// Remove handler
func (t *TickManager) Remove(id uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h, ok := t.ticker[id]
	if ok {
		h.cancel()
		delete(t.ticker, id)
	}
}

// Stop all handlers
func (t *TickManager) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, h := range t.ticker {
		h.cancel()
	}
}

// GetHandlers return all handlers
func (t *TickManager) GetHandlers() map[uint64]*TickHandler {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := map[uint64]*TickHandler{}
	for id, h := range t.ticker {
		result[id] = h
	}

	return result
}

// Run all tickers
func (t *TickManager) Run() {
	tickers := t.GetHandlers()

	var wg sync.WaitGroup
	wg.Add(len(tickers))

	for id, h := range tickers {
		go func(id uint64, h *TickHandler) {
			defer wg.Done()

			t.run(id, h)
		}(id, h)
	}

	wg.Wait()
}

// CountTicksInProgress return count of ticks in progress
func (t *TickManager) CountTicksInProgress() int32 {
	return atomic.LoadInt32(&t.count)
}

func (t *TickManager) run(id uint64, h *TickHandler) {
	tickTimer := time.NewTicker(h.interval)
	tickTimerDebug := time.NewTicker(time.Second)

	for {
		select {
		case <-h.ctx.Done():
			tickTimer.Stop()
			t.Remove(id)
			slog.Default().Info("Stop tick handler")

			return
		case <-tickTimerDebug.C:
			c := t.CountTicksInProgress()
			if c > maxDelayedTicks {
				t.Stop()
				slog.Default().Info("Stop tick manager", "count", c)
			}

			slog.Default().Debug("Tick debug", "in progress", t.CountTicksInProgress())
		case <-tickTimer.C:
			atomic.AddInt32(&t.count, 1)

			go func() {
				ctxTimeout, cancelTimeout := context.WithTimeout(h.ctx, h.interval)

				defer cancelTimeout()
				defer atomic.AddInt32(&t.count, -1)

				st := time.Now()

				h.handler.Tick(ctxTimeout)
				slog.Default().Debug("Tick finish", "duration", time.Since(st))
			}()
		}
	}
}
