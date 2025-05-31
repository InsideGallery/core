//go:build integration
// +build integration

package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/InsideGallery/core/memory/registry"
	"github.com/InsideGallery/core/queue/nats/client"
	"github.com/InsideGallery/core/queue/nats/publisher"
	"github.com/InsideGallery/core/queue/nats/subscriber"
	"github.com/InsideGallery/core/testutils"

	"github.com/nats-io/nats.go"
)

func createClient(natsClient *client.Client, clientName string) (*Client, string) {
	cl := NewClient(publisher.New(natsClient), "proxy_service_subject", func() string {
		return clientName
	})
	// We should retrieve subject to subscribe, but server can be unavailable
	var proxiesSubject string
	var err error
	for {
		proxiesSubject, err = cl.Subscribe(context.Background(), "subject_1")
		if err != nil {
			continue
		}
		break
	}
	return cl, proxiesSubject
}

func TestProxyServer(t *testing.T) {
	natsClient, err := client.Default(context.Background(), nil, "NATS")
	testutils.Equal(t, err, nil)

	pool := subscriber.NewSubscriber(natsClient)

	// Init server, usually it's ready in half of second
	srv := NewServer(publisher.New(natsClient), "proxy_service_subject", []string{"subject_1", "subject_2"})
	err = srv.Process(pool)
	testutils.Equal(t, err, nil)

	// Process ping service to ping subscribers every 2 seconds
	pg := NewPing(publisher.New(natsClient), time.Second)
	go pg.Service(context.Background(), srv)
	cl, proxiesSubject := createClient(natsClient, "test_client_1")
	_, proxiesSubject2 := createClient(natsClient, "test_client_1")

	slog.Default().Info("Subject get from server", "subject", proxiesSubject)

	testutils.Equal(t, proxiesSubject, proxiesSubject2) // should be equal by same client
	testutils.Equal(t, cl.PongListener(pool), nil)
	client2, proxiesSubject2 := createClient(natsClient, "test_client_2")

	testutils.NotEqual(t, proxiesSubject, proxiesSubject2) // should be different by different client
	testutils.Equal(t, client2.PongListener(pool), nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)
	// Start listen
	pool.Subscribe(proxiesSubject, "queue", func(ctx context.Context, msg *nats.Msg) error {
		fmt.Println("RECEIVE MESSAGE on client 1", string(msg.Data))
		ch1 <- registry.Nothing
		return nil
	})
	pool.Subscribe(proxiesSubject2, "queue", func(ctx context.Context, msg *nats.Msg) error {
		fmt.Println("RECEIVE MESSAGE on client 2", string(msg.Data))
		ch2 <- registry.Nothing
		return nil
	})

	p := publisher.New(natsClient)
	// Try to send message to expected service 1
	err = p.Publish("subject_1", []byte("test message"), HeaderID, "test_id")
	testutils.Equal(t, err, nil)

	// Try to send message to expected service 2
	err = p.Publish("subject_1", []byte("test message"), HeaderID, "test_id3")
	testutils.Equal(t, err, nil)

	// Validate response
	select {
	case <-ctx.Done():
		t.Fatalf("Error by timeout, client 1 does not receive message")
	case <-ch1:
	}
	select {
	case <-ctx.Done():
		t.Fatalf("Error by timeout, client 2 does not receive message")
	case <-ch2:
	}

	// Try to force ping client 1
	err = pg.Ping(srv, "test_client_1")
	testutils.Equal(t, err, nil)
	// Try to force ping client 2
	err = pg.Ping(srv, "test_client_2")
	testutils.Equal(t, err, nil)
}
