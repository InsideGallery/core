package subscriber

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/multiproc/worker"
)

const (
	maximumBufferSize = 1024

	poolSizeValidate = 100 * time.Millisecond
	idleTimeout      = 30 * time.Second
)

type Subscription struct {
	wpool  *worker.WorkersPool[*nats.Msg]
	ctx    context.Context
	cancel func()
	Client
	*nats.Subscription
	subject, queue string
}

func NewSubscription(c Client, subject, queue string) (*Subscription, error) {
	sub, err := c.QueueSubscribeSync(subject, queue)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(c.Context())

	return &Subscription{
		ctx:          ctx,
		cancel:       cancel,
		Client:       c,
		Subscription: sub,
		subject:      subject,
		queue:        queue,
		wpool:        worker.NewWorkersPool[*nats.Msg](c.Context()),
	}, nil
}

func (s *Subscription) Process(
	ctx context.Context,
	buffer int,
	timeout time.Duration,
	handler MsgHandler,
) error {
	err := s.setupMetrics()
	if err != nil {
		return err
	}

	ch := make(chan *nats.Msg, maximumBufferSize)

	// Read messages
	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.ctx.Done():
				return
			default:
				msg, err := s.Subscription.NextMsg(timeout) // Read with timeout
				if err != nil {
					s.Client.Logger().Debug("Error getting next message with timeout",
						"err", err,
						"subject", s.subject,
						"queue", s.queue,
					)

					continue
				}

				ch <- msg
			}
		}
	}()

	go func() {
		t := time.NewTicker(poolSizeValidate)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if len(ch) > 0 && s.wpool.Size() < s.Config().MaxConcurrentSize {
					s.Client.Logger().Info("Increase pool size", "pending", len(ch), "wpool", s.wpool.Size())
					s.wpool.Execute(func(ctx context.Context) error {
						return s.wpool.TemporalWorker(ctx, idleTimeout, func() {
							s.Client.Logger().Info("Decrease pool size", "wpool", s.wpool.Size()-1)
						}, ch, handler)
					})
				}
			}
		}
	}()

	// Setup default buffer
	for i := 0; i < buffer; i++ {
		s.wpool.Execute(func(ctx context.Context) error {
			return s.wpool.PersistentWorker(ctx, ch, handler)
		})
	}

	return s.wpool.Wait()
}

func (s *Subscription) Stop() error {
	err := s.Subscription.Drain()
	if err != nil {
		return err
	}

	s.wpool.Stop()

	s.cancel()

	return nil
}
