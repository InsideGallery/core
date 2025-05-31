package subscriber

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/nats/client"
)

const (
	HeaderConsumerError = "Error"
)

// WithResponseOnError wraps handler where if handler returns a non-nil error,
// the msg is then responded to with said error's string in the
// HeaderConsumerError header. Such message then causes nats publisher to return
// the error when reading response.
//
// Note: this wrapper should NOT be used for observer-like handlers that do not
// send success responses. If subscribed to the same subject with an actual
// responder, the latter's response can potentially get snubbed.
func WithResponseOnError(logger client.Logger, handler MsgHandler) MsgHandler {
	if logger == nil {
		logger = client.StubLogger{}
	}

	return func(ctx context.Context, msg *nats.Msg) error {
		err := handler(ctx, msg)
		if err != nil && len(msg.Reply) > 0 {
			responseMsg := nats.NewMsg(msg.Reply)
			responseMsg.Header.Set(HeaderConsumerError, err.Error())

			if respErr := msg.RespondMsg(responseMsg); respErr != nil {
				logger.Error("Message error response failed", "err", err)
			}
		}

		return err
	}
}
