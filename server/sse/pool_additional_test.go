package sse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPoolLifecycle(t *testing.T) {
	cases := []struct {
		name       string
		bufferSize int
	}{
		{
			name:       "explicit buffer",
			bufferSize: 1,
		},
		{
			name:       "default buffer",
			bufferSize: 0,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			pool := NewPool(test.bufferSize)

			if pool.Connections() != 0 {
				t.Fatalf("connections = %d, want 0", pool.Connections())
			}

			first := pool.Add("alice")
			second := pool.Add("bob")

			if pool.Connections() != 2 {
				t.Fatalf("connections = %d, want 2", pool.Connections())
			}

			users := pool.GetAllConnectedUsers()
			if !users.Contains("alice") || !users.Contains("bob") {
				t.Fatalf("connected users missing expected entries: %#v", users)
			}

			message := NewMessage("notice", "hello")
			if err := pool.Send("alice", message); err != nil {
				t.Fatalf("send: %v", err)
			}

			if got := <-first; got.Event != message.Event {
				t.Fatalf("message event = %q, want %q", got.Event, message.Event)
			}

			pool.SendToAll(message)

			if got := <-first; got.Event != message.Event {
				t.Fatalf("broadcast first event = %q, want %q", got.Event, message.Event)
			}

			if got := <-second; got.Event != message.Event {
				t.Fatalf("broadcast second event = %q, want %q", got.Event, message.Event)
			}

			if err := pool.Send("missing", message); !errors.Is(err, ErrNotFoundConnectedUser) {
				t.Fatalf("missing send err = %v, want %v", err, ErrNotFoundConnectedUser)
			}

			pool.Remove("alice")

			if pool.Connections() != 1 {
				t.Fatalf("connections after remove = %d, want 1", pool.Connections())
			}

			pool.StopAll()

			if pool.Connections() != 0 {
				t.Fatalf("connections after stop = %d, want 0", pool.Connections())
			}
		})
	}
}

func TestPoolHandler(t *testing.T) {
	cases := []struct {
		name      string
		withUser  bool
		wantUsers int
	}{
		{
			name:     "missing user id returns without connection",
			withUser: false,
		},
		{
			name:     "valid user is added and removed",
			withUser: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			pool := NewPool(1)
			recorder := httptest.NewRecorder()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			if test.withUser {
				ctx = context.WithValue(ctx, ContextUserID, "alice")
			}

			request := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
			pool.Handler(recorder, request)

			if pool.Connections() != test.wantUsers {
				t.Fatalf("connections = %d, want %d", pool.Connections(), test.wantUsers)
			}
		})
	}
}
