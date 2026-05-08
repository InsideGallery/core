package storage

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"

	coreredis "github.com/InsideGallery/core/db/redis"
)

func TestNewRedisWithOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options RedisOptions
		wantKey string
	}{
		{
			name: "uses core redis connection",
			options: RedisOptions{
				Context: context.Background(),
				Key:     "proxy",
				Connection: coreredis.NewRedisClient( //nolint:staticcheck // verifies compatibility adapter input
					&coreredis.ConnectionConfig{Host: "127.0.0.1", Port: "6379"},
				),
			},
			wantKey: "proxy",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := NewRedisWithOptions(test.options)
			t.Cleanup(func() {
				if test.options.Connection != nil {
					if err := test.options.Connection.Stop(); err != nil {
						t.Errorf("stop redis client: %v", err)
					}
				}
			})

			if got.key != test.wantKey {
				t.Fatalf("key = %q, want %q", got.key, test.wantKey)
			}

			if got.ctx == nil {
				t.Fatal("ctx = nil, want non-nil")
			}
		})
	}
}

func TestRedisLockOrWaitReturnsErrorAfterRetryExhaustion(t *testing.T) {
	t.Parallel()

	lockErr := errors.New("lock unavailable")
	cases := []struct {
		name string
		pool failingPool
	}{
		{
			name: "failing redis pool",
			pool: failingPool{err: lockErr},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rs := redsync.New(test.pool)
			storage := &Redis{
				ctx:   context.Background(),
				mutex: rs.NewMutex(GetMutexName("test"), redsync.WithTries(1), redsync.WithRetryDelay(0)),
			}

			err := storage.LockOrWait()
			if err == nil {
				t.Fatal("expected lock error")
			}

			if !errors.Is(err, lockErr) {
				t.Fatalf("err = %v, want wrapped %v", err, lockErr)
			}

			if !strings.Contains(err.Error(), "lock failed after 10 attempts") {
				t.Fatalf("err = %v, want retry exhaustion context", err)
			}
		})
	}
}

type failingPool struct {
	err error
}

func (p failingPool) Get(context.Context) (redsyncredis.Conn, error) { //nolint:ireturn // implements redsync redis.Pool
	return nil, p.err
}
