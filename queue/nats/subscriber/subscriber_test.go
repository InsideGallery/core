//go:build integration
// +build integration

package subscriber

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/InsideGallery/core/multiproc/worker"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/testutils"

	"github.com/nats-io/nats.go"
)

func TestPublisher(t *testing.T) {
	t.Skip()
	conn, err := client.Default(context.TODO(), slog.Default())
	testutils.Equal(t, err, nil)

	subject1 := "core_subject_1"

	total := 1000
	var timeouts uint64
	var success uint64

	go func() {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()

		for range tk.C {
			fmt.Println("TOTAL:", total, "SUCCESS:", success, "TIMEOUTS:", timeouts)
		}
	}()

	st := time.Now()
	pool := worker.NewPool(context.TODO())
	for i := 0; i < total; i++ {
		pool.Execute(func(ctx context.Context) error {
			st := time.Now()
			resp, err := conn.Conn().Request(subject1, []byte(strconv.Itoa(i)), 5*time.Second)
			if err != nil {
				atomic.AddUint64(&timeouts, 1)
				slog.Error("Error publish message ", "err", err)
				return nil
			}

			atomic.AddUint64(&success, 1)
			fmt.Println(string(resp.Data), fmt.Sprint(time.Since(st).String()))
			return nil
		})
		time.Sleep(time.Millisecond * 10)
	}

	err = pool.Wait()

	fmt.Println("TAKE", time.Since(st).String(), "TOTAL:", total, "SUCCESS:", success, "TIMEOUTS:", timeouts)

	testutils.Equal(t, err, nil)
}

func TestSubscriber1(t *testing.T) {
	t.Skip()
	conn, err := client.Default(context.TODO(), slog.Default())
	testutils.Equal(t, err, nil)

	s := NewSubscriber(conn)

	ch := make(chan struct{})

	subject1 := "core_subject_1"

	s.Subscribe(subject1, "test", func(ctx context.Context, msg *nats.Msg) error {
		//slog.Info("Received message", "data", string(msg.Data), "g", runtime.NumGoroutine())
		time.Sleep(time.Millisecond * 500)
		err = msg.Respond(msg.Data)
		if err != nil {
			slog.Error("Error sending response ", "err", err)
		}
		return nil
	})

	ch <- struct{}{}
	err = s.Wait()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, len(s.subs.GetMap()), 0)
}

func TestSubscriber(t *testing.T) {
	conn, err := client.Default(context.TODO(), slog.Default())
	testutils.Equal(t, err, nil)

	s := NewSubscriber(conn)

	err = conn.Conn().Flush()
	testutils.Equal(t, err, nil)
	ch := make(chan struct{})

	subject1 := "core_subject_1"
	subject2 := "core_subject_2"
	subject3 := "core_subject_3"

	go func() {
		<-ch
		time.Sleep(30 * time.Millisecond)

		defer func() {
			err := s.Close()
			if err != nil {
				slog.Error("Error stop subscriber", "err", err)
			}

			err = conn.Close()
			if err != nil {
				slog.Error("Error closing connection", "err", err)
			}
		}()

		for i := 0; i < 10; i++ {
			err := conn.Conn().Publish(subject1, []byte(strconv.Itoa(i)))
			if err != nil {
				slog.Error("Error publish message ", "err", err)
				continue
			}

			resp, err := conn.Conn().Request(subject2, []byte(strconv.Itoa(i)), 30*time.Millisecond)
			if err != nil {
				slog.Error("Error publish message ", "err", err)
				continue
			}

			testutils.Equal(t, string(resp.Data), strconv.Itoa(i))

			resp2, err := conn.Conn().Request(subject3, []byte(strconv.Itoa(i)), 30*time.Millisecond)
			if err != nil {
				slog.Error("Error publish message ", "err", err)
				continue
			}

			testutils.Equal(t, resp2.Header.Get(HeaderConsumerError), "error send response")
			time.Sleep(time.Millisecond * 2)
		}
	}()

	s.Subscribe(subject1, "test", func(ctx context.Context, msg *nats.Msg) error {
		slog.Info("Received message", "data", string(msg.Data))
		time.Sleep(time.Millisecond * 2)
		return nil
	})

	s.Subscribe(subject2, "test", func(ctx context.Context, msg *nats.Msg) error {
		slog.Info("Received message", "data", string(msg.Data))
		time.Sleep(time.Millisecond * 2)
		err := msg.Respond(msg.Data)
		if err != nil {
			slog.Error("Error sending response ", "err", err)
		}
		return nil
	})

	s.Subscribe(subject3, "test", WithResponseOnError(slog.Default(), func(ctx context.Context, msg *nats.Msg) error {
		slog.Info("Received message", "data", string(msg.Data))
		time.Sleep(time.Millisecond * 2)
		return errors.New("error send response")
	}))

	ch <- struct{}{}
	err = s.Wait()
	testutils.Equal(t, err, nil)

	testutils.Equal(t, len(s.subs.GetMap()), 0)
}
