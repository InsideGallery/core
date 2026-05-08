package nats

import (
	"context"
	"errors"
	"testing"
)

func TestPublisherContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		operation func(context.Context, *PublisherAdapter) error
		wantErr   error
	}{
		{
			name: "publish missing publisher",
			operation: func(ctx context.Context, adapter *PublisherAdapter) error {
				_, err := adapter.Publish(ctx, PublishOptions{Message: Message{Subject: "events"}})
				return err
			},
			wantErr: ErrPublisherNotSet,
		},
		{
			name: "request missing publisher",
			operation: func(ctx context.Context, adapter *PublisherAdapter) error {
				_, err := adapter.Request(ctx, RequestOptions{Message: Message{Subject: "events"}})
				return err
			},
			wantErr: ErrPublisherNotSet,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var _ Publisher = (*PublisherAdapter)(nil)
			var _ Subscriber = (*SubscriberAdapter)(nil)

			err := test.operation(context.Background(), NewPublisher(nil))
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("operation() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestSubscriberAdapterMissingDependency(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		operation func(context.Context, *SubscriberAdapter) error
		wantErr   error
	}{
		{
			name: "subscribe missing subscriber",
			operation: func(ctx context.Context, adapter *SubscriberAdapter) error {
				_, err := adapter.Subscribe(ctx, SubscribeOptions{Subject: "events"}, func(context.Context, Message) error {
					return nil
				})

				return err
			},
			wantErr: ErrSubscriberNotSet,
		},
		{
			name: "close missing subscriber",
			operation: func(ctx context.Context, adapter *SubscriberAdapter) error {
				return adapter.Close(ctx)
			},
			wantErr: ErrSubscriberNotSet,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.operation(context.Background(), NewSubscriber(nil))
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("operation() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestFlattenHeaders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		headers Headers
		want    []string
	}{
		{
			name:    "empty",
			headers: nil,
			want:    nil,
		},
		{
			name: "sorted",
			headers: Headers{
				"b": {"2"},
				"a": {"1"},
			},
			want: []string{"a", "1", "b", "2"},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := flattenHeaders(test.headers)
			if len(got) != len(test.want) {
				t.Fatalf("len(flattenHeaders()) = %d, want %d", len(got), len(test.want))
			}

			for i := range test.want {
				if got[i] != test.want[i] {
					t.Fatalf("flattenHeaders()[%d] = %q, want %q", i, got[i], test.want[i])
				}
			}
		})
	}
}
