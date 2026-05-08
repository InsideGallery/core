package subscriber

import (
	"context"
	"errors"
	"time"

	natsgo "github.com/nats-io/nats.go"
)

// ErrSubscriberNotSet reports a nil subscriber dependency.
var ErrSubscriberNotSet = errors.New("subscriber is not set")

// ErrHandlerNotSet reports a nil message handler.
var ErrHandlerNotSet = errors.New("handler is not set")

// ErrSubjectNotSet reports a missing subscription subject.
var ErrSubjectNotSet = errors.New("subject is not set")

// Headers is the core-owned subscriber message header shape.
type Headers map[string][]string

// Message is the core-owned subscriber message shape.
type Message struct {
	Subject string
	Reply   string
	Data    []byte
	Header  Headers
}

// MessageHandler handles a message without exposing the NATS SDK message type.
type MessageHandler func(ctx context.Context, msg Message) error

// SubscribeOptions is the core-owned input for NATS queue subscriptions.
type SubscribeOptions struct {
	Subject string
	Queue   string
	Buffer  int
	Timeout time.Duration
}

// SubscribeResult reports the registered subscription identity.
type SubscribeResult struct {
	Subject string
	Queue   string
}

// MessageSubscriber is the core-owned subscriber contract for new consumers.
type MessageSubscriber interface {
	SubscribeMessage(ctx context.Context, options SubscribeOptions, handler MessageHandler) (SubscribeResult, error)
	Close() error
	Wait() error
}

// SubscribeMessage registers a message handler through core-owned inputs.
func (s *Subscriber) SubscribeMessage(
	ctx context.Context,
	options SubscribeOptions,
	handler MessageHandler,
) (SubscribeResult, error) {
	if s == nil {
		return SubscribeResult{}, ErrSubscriberNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return SubscribeResult{}, err
	}

	if options.Subject == "" {
		return SubscribeResult{}, ErrSubjectNotSet
	}

	if handler == nil {
		return SubscribeResult{}, ErrHandlerNotSet
	}

	buffer := options.Buffer
	if buffer <= 0 {
		buffer = s.Config().GetConcurrentSize()
	}

	timeout := options.Timeout
	if timeout <= 0 {
		timeout = s.Config().GetReadTimeout()
	}

	s.SubscribeWithParameters(
		buffer,
		timeout,
		options.Subject,
		options.Queue,
		func(ctx context.Context, msg *natsgo.Msg) error {
			return handler(ctx, newMessage(msg))
		},
	)

	return SubscribeResult{
		Subject: options.Subject,
		Queue:   options.Queue,
	}, nil
}

func newMessage(msg *natsgo.Msg) Message {
	if msg == nil {
		return Message{}
	}

	return Message{
		Subject: msg.Subject,
		Reply:   msg.Reply,
		Data:    append([]byte(nil), msg.Data...),
		Header:  cloneHeaders(msg.Header),
	}
}

func cloneHeaders(headers map[string][]string) Headers {
	if len(headers) == 0 {
		return nil
	}

	cloned := make(Headers, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}
