package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/nats/subscriber"
)

type Recovery struct{}

func NewRecovery() *Recovery {
	return &Recovery{}
}

func (t *Recovery) Call(next subscriber.MsgHandler) subscriber.MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				slog.Default().Error("panic recovered",
					"error", fmt.Sprintf("%v", r),
					"stack", string(stack),
				)
			}
		}()

		return next(ctx, msg)
	}
}
