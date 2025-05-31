//go:build integration
// +build integration

package publisher

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/testutils"
)

func TestPublisher(t *testing.T) {
	conn, err := client.Default(context.TODO(), slog.Default())
	testutils.Equal(t, err, nil)

	subject1 := "core_subject_1"
	subject2 := "core_subject_2"

	_, err = conn.Conn().Subscribe(subject1, func(msg *nats.Msg) {
		err := msg.Respond([]byte("result"))
		testutils.Equal(t, err, nil)
	})
	testutils.Equal(t, err, nil)

	_, err = conn.Conn().Subscribe(subject2, func(_ *nats.Msg) {})
	testutils.Equal(t, err, nil)

	time.Sleep(500 * time.Millisecond)

	p := New(conn)
	res, err := p.Requester(subject1, []byte("test message"), 100*time.Millisecond)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, string(res), "result")
	err = p.Publish(subject2, []byte("test message2"))
	testutils.Equal(t, err, nil)
}
