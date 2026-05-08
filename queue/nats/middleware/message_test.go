package middleware

import (
	"context"
	"testing"

	corenats "github.com/InsideGallery/core/queue/nats"
)

func TestMessageMiddleware(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		chains []MessageChain
		want   string
	}{
		{
			name: "wraps handler",
			chains: []MessageChain{
				func(next corenats.MessageHandler) corenats.MessageHandler {
					return func(ctx context.Context, msg corenats.Message) error {
						msg.Data = append([]byte("chain:"), msg.Data...)

						return next(ctx, msg)
					}
				},
			},
			want: "chain:handler",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var got string
			middleware := NewMessageMiddleware(test.chains...)
			handler := middleware.Then(func(_ context.Context, msg corenats.Message) error {
				got = string(msg.Data)

				return nil
			})

			if err := handler(context.Background(), corenats.Message{Data: []byte("handler")}); err != nil {
				t.Fatalf("handler: %v", err)
			}

			if got != test.want {
				t.Fatalf("message = %q, want %q", got, test.want)
			}
		})
	}
}

func TestGetMessageChainsIncludesRecovery(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want int
	}{
		{
			name: "default recovery",
			want: 1,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := GetMessageChains()
			if len(got) != test.want {
				t.Fatalf("len(GetMessageChains()) = %d, want %d", len(got), test.want)
			}
		})
	}
}
