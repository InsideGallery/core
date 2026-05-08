// Package nats provides NATS publishing and subscription helpers.
//
// New code should use core-owned message and boundary contracts at consumer
// edges:
//
//	import "github.com/InsideGallery/core/queue/nats"
//
//	publisher := nats.NewPublisher(legacyPublisher)
//	result, err := publisher.Publish(ctx, nats.PublishOptions{
//		Message: nats.Message{Subject: "events"},
//	})
//
// Prefer Publisher, Subscriber, Message, Headers, PublishOptions,
// RequestOptions, SubscribeOptions, and their result types where application code
// should not depend on the NATS SDK message type.
//
// Compatibility: the queue/nats/client, publisher, and subscriber sub-packages
// remain available for existing SDK-shaped call sites. Prefer ConnectClient and
// the adapters in this package for new code.
package nats

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/InsideGallery/core/queue/nats/publisher"
	legacysubscriber "github.com/InsideGallery/core/queue/nats/subscriber"
)

const headerKeyValuePairSize = 2

// ErrMissingSubject reports a message without a subject.
var ErrMissingSubject = errors.New("subject is not set")

// ErrPublisherNotSet reports a missing publisher dependency.
var ErrPublisherNotSet = errors.New("publisher is not set")

// ErrSubscriberNotSet reports a missing subscriber dependency.
var ErrSubscriberNotSet = errors.New("subscriber is not set")

// ErrHandlerNotSet reports a missing subscriber handler.
var ErrHandlerNotSet = errors.New("handler is not set")

// Headers is the core-owned queue header shape.
type Headers map[string][]string

// Message is the core-owned queue message shape.
type Message struct {
	Subject string
	Reply   string
	Data    []byte
	Header  Headers
}

// PublishOptions is the core-owned input for publishing a message.
type PublishOptions struct {
	Message Message
	Timeout time.Duration
}

// RequestOptions is the core-owned input for request/reply messaging.
type RequestOptions struct {
	Message Message
}

// SubscribeOptions is the core-owned input for NATS queue subscriptions.
type SubscribeOptions struct {
	Subject string
	Queue   string
	Buffer  int
	Timeout time.Duration
}

// PublishResult is the core-owned publish result.
type PublishResult struct {
	Subject string
	Bytes   int
}

// RequestResult is the core-owned request/reply result.
type RequestResult struct {
	Message Message
}

// SubscribeResult reports the registered subscription identity.
type SubscribeResult struct {
	Subject string
	Queue   string
}

// MessageHandler handles a message without exposing the NATS SDK message type.
type MessageHandler func(ctx context.Context, msg Message) error

// Publisher is the core-owned NATS publishing contract for new consumers.
type Publisher interface {
	Publish(ctx context.Context, options PublishOptions) (PublishResult, error)
	Request(ctx context.Context, options RequestOptions) (RequestResult, error)
}

// Subscriber is the core-owned NATS subscriber contract for new consumers.
type Subscriber interface {
	Subscribe(ctx context.Context, options SubscribeOptions, handler MessageHandler) (SubscribeResult, error)
	Close(ctx context.Context) error
	Wait(ctx context.Context) error
}

// PublisherAdapter adapts the legacy NATS publisher to the core-owned Publisher contract.
type PublisherAdapter struct {
	publisher *publisher.Publisher
}

// SubscriberAdapter adapts the legacy NATS subscriber to the core-owned Subscriber contract.
type SubscriberAdapter struct {
	subscriber *legacysubscriber.Subscriber
}

// NewPublisher wraps a legacy NATS publisher with the core-owned Publisher contract.
func NewPublisher(p *publisher.Publisher) *PublisherAdapter {
	return &PublisherAdapter{publisher: p}
}

// NewSubscriber wraps a legacy NATS subscriber with the core-owned Subscriber contract.
func NewSubscriber(s *legacysubscriber.Subscriber) *SubscriberAdapter {
	return &SubscriberAdapter{subscriber: s}
}

// Publish publishes a message through the core-owned Publisher contract.
func (p *PublisherAdapter) Publish(ctx context.Context, options PublishOptions) (PublishResult, error) {
	if p == nil || p.publisher == nil {
		return PublishResult{}, ErrPublisherNotSet
	}

	if options.Message.Subject == "" {
		return PublishResult{}, ErrMissingSubject
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if options.Timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	if err := p.publisher.PublishWithContext(
		ctx,
		options.Message.Subject,
		options.Message.Data,
		flattenHeaders(options.Message.Header)...,
	); err != nil {
		return PublishResult{}, fmt.Errorf("nats publish %q: %w", options.Message.Subject, err)
	}

	return PublishResult{
		Subject: options.Message.Subject,
		Bytes:   len(options.Message.Data),
	}, nil
}

// Request sends a request and returns a core-owned response message.
func (p *PublisherAdapter) Request(ctx context.Context, options RequestOptions) (RequestResult, error) {
	if p == nil || p.publisher == nil {
		return RequestResult{}, ErrPublisherNotSet
	}

	if options.Message.Subject == "" {
		return RequestResult{}, ErrMissingSubject
	}

	if ctx == nil {
		ctx = context.Background()
	}

	data, err := p.publisher.RequesterWithContext(
		ctx,
		options.Message.Subject,
		options.Message.Data,
		flattenHeaders(options.Message.Header)...,
	)
	if err != nil {
		return RequestResult{}, fmt.Errorf("nats request %q: %w", options.Message.Subject, err)
	}

	return RequestResult{
		Message: Message{
			Subject: options.Message.Subject,
			Data:    data,
		},
	}, nil
}

// Subscribe registers a queue subscriber through the core-owned Subscriber contract.
func (s *SubscriberAdapter) Subscribe(
	ctx context.Context,
	options SubscribeOptions,
	handler MessageHandler,
) (SubscribeResult, error) {
	if s == nil || s.subscriber == nil {
		return SubscribeResult{}, ErrSubscriberNotSet
	}

	if handler == nil {
		return SubscribeResult{}, ErrHandlerNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if options.Subject == "" {
		return SubscribeResult{}, ErrMissingSubject
	}

	result, err := s.subscriber.SubscribeMessage(ctx, legacysubscriber.SubscribeOptions{
		Subject: options.Subject,
		Queue:   options.Queue,
		Buffer:  options.Buffer,
		Timeout: options.Timeout,
	}, func(ctx context.Context, msg legacysubscriber.Message) error {
		return handler(ctx, messageFromSubscriber(msg))
	})
	if err != nil {
		return SubscribeResult{}, fmt.Errorf("nats subscribe %q: %w", options.Subject, err)
	}

	return SubscribeResult{
		Subject: result.Subject,
		Queue:   result.Queue,
	}, nil
}

// Close closes the wrapped subscriber.
func (s *SubscriberAdapter) Close(ctx context.Context) error {
	if s == nil || s.subscriber == nil {
		return ErrSubscriberNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := s.subscriber.Close(); err != nil {
		return fmt.Errorf("nats subscriber close: %w", err)
	}

	return nil
}

// Wait waits until the wrapped subscriber stops.
func (s *SubscriberAdapter) Wait(ctx context.Context) error {
	if s == nil || s.subscriber == nil {
		return ErrSubscriberNotSet
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := s.subscriber.Wait(); err != nil {
		return fmt.Errorf("nats subscriber wait: %w", err)
	}

	return nil
}

func flattenHeaders(headers Headers) []string {
	if len(headers) == 0 {
		return nil
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	values := make([]string, 0, len(headers)*headerKeyValuePairSize)
	for _, key := range keys {
		for _, value := range headers[key] {
			values = append(values, key, value)
		}
	}

	return values
}

func messageFromSubscriber(msg legacysubscriber.Message) Message {
	return Message{
		Subject: msg.Subject,
		Reply:   msg.Reply,
		Data:    append([]byte(nil), msg.Data...),
		Header:  headersFromSubscriber(msg.Header),
	}
}

func headersFromSubscriber(headers legacysubscriber.Headers) Headers {
	if len(headers) == 0 {
		return nil
	}

	cloned := make(Headers, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}
