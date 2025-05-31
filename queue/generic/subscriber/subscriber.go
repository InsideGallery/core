package subscriber

import (
	"context"
	"errors"
	"runtime"
	"time"

	memory "github.com/InsideGallery/core/memory/utils"
	"github.com/InsideGallery/core/multiproc/worker"
	"github.com/InsideGallery/core/queue/generic/subscriber/interfaces"
)

func GetConcurrentSize(cfg interfaces.Config) int {
	size := cfg.GetConcurrentSize()
	if size <= 0 {
		return runtime.NumCPU()
	}

	return size
}

type Subscriber struct {
	interfaces.Client
	*worker.Pool

	subs *memory.SafeMap[string, *Subscription]
}

func NewSubscriber(c interfaces.Client) *Subscriber {
	return &Subscriber{
		Client: c,
		Pool:   worker.NewPool(c.Context()),
		subs:   memory.NewSafeMap[string, *Subscription](nil),
	}
}

func (s *Subscriber) GetSubs() *memory.SafeMap[string, *Subscription] {
	return s.subs
}

func (s *Subscriber) Subscribe(subject, queue string, handler interfaces.MsgHandler) {
	s.SubscribeWithParameters(GetConcurrentSize(s.Config()), s.Config().GetReadTimeout(), subject, queue, handler)
}

func (s *Subscriber) SubscribeWithParameters(
	buffer int, timeout time.Duration, subject, queue string, handler interfaces.MsgHandler,
) {
	if handler == nil {
		return
	}

	if buffer <= 0 {
		buffer = 1
	}

	s.Pool.Execute(func(ctx context.Context) error {
		sub, err := NewSubscription(s, subject, queue)
		if err != nil {
			return err
		}

		s.subs.Set(subject+":"+queue, sub)

		return sub.Process(ctx, buffer, timeout, handler)
	})
}

func (s *Subscriber) Close() error {
	var errs []error
	for key, sub := range s.subs.GetMap() {
		errs = append(errs, sub.Stop())

		s.subs.Remove(key)
	}

	return errors.Join(errs...)
}

func (s *Subscriber) Get(subject, queue string) *Subscription {
	sub, _ := s.subs.Get(subject + ":" + queue)

	return sub
}

func (s *Subscriber) Wait() error {
	return s.Pool.Wait()
}

func (s *Subscriber) ForceClose() {
	s.Pool.Close()
}
