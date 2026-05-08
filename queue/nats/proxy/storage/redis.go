package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"

	coreredis "github.com/InsideGallery/core/db/redis"
)

const (
	MutexPrefix   = "mutex_"
	DefaultPrefix = "default:"

	lockAttempts = 10
)

func GetMutexName(key string) string {
	return MutexPrefix + key
}

type Redis struct {
	ctx   context.Context
	conn  *redis.Client
	mutex *redsync.Mutex
	key   string
}

// RedisOptions is the core-owned input for Redis-backed proxy storage.
type RedisOptions struct {
	Context    context.Context
	Key        string
	Connection *coreredis.Connection //nolint:staticcheck // adapts the core Redis compatibility wrapper
}

// NewRedis creates Redis-backed proxy storage from a Redis SDK client.
//
// Deprecated: use NewRedisWithOptions with db/redis.Connection for new code.
func NewRedis(ctx context.Context, key string, conn *redis.Client) *Redis {
	pool := goredis.NewPool(conn)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(GetMutexName(key))

	return &Redis{
		ctx:   ctx,
		key:   key,
		conn:  conn,
		mutex: mutex,
	}
}

// NewRedisWithOptions creates Redis-backed proxy storage without exposing Redis SDK types.
func NewRedisWithOptions(options RedisOptions) *Redis {
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var conn *redis.Client
	if options.Connection != nil {
		conn = options.Connection.Client
	}

	return NewRedis(ctx, options.Key, conn)
}

func (s *Redis) LockOrWait() error {
	var err error

	for i := 0; i < lockAttempts; i++ {
		err = s.mutex.LockContext(s.ctx)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("lock failed after %d attempts: %w", lockAttempts, err)
}

func (s *Redis) TryLock() {
	err := s.mutex.TryLockContext(s.ctx)
	if err != nil {
		slog.Default().Debug("Error lock in redis", "err", err)
	}
}

func (s *Redis) TryUnlock() {
	_, err := s.mutex.UnlockContext(s.ctx)
	if err != nil {
		slog.Default().Debug("Error unlock in redis", "err", err)
	}
}

func (s *Redis) Add(group string, id string) error {
	if err := s.LockOrWait(); err != nil {
		return fmt.Errorf("add lock: %w", err)
	}

	defer s.TryUnlock()

	score := float64(time.Now().Unix())

	err := s.conn.ZAdd(s.ctx, DefaultPrefix+group, redis.Z{
		Score:  score,
		Member: id,
	}).Err()
	if err != nil {
		return fmt.Errorf("error set id for group: %w", err)
	}

	err = s.conn.ZAdd(s.ctx, DefaultPrefix+id, redis.Z{
		Score:  score,
		Member: group,
	}).Err()
	if err != nil {
		return fmt.Errorf("error set group for id: %w", err)
	}

	err = s.conn.ZAdd(s.ctx, DefaultPrefix+s.key, redis.Z{
		Score:  score,
		Member: id,
	}).Err()
	if err != nil {
		return fmt.Errorf("error set id for global key: %w", err)
	}

	return nil
}

func (s *Redis) Delete(group string, id string) error {
	if err := s.LockOrWait(); err != nil {
		return fmt.Errorf("delete lock: %w", err)
	}

	defer s.TryUnlock()

	err := s.conn.ZRem(s.ctx, DefaultPrefix+s.key, id).Err()
	if err != nil {
		return err
	}

	err = s.conn.ZRem(s.ctx, DefaultPrefix+group, id).Err()
	if err != nil {
		return err
	}

	return s.conn.ZRem(s.ctx, DefaultPrefix+id, group).Err()
}

func (s *Redis) DeleteByID(id string) error {
	if err := s.LockOrWait(); err != nil {
		return fmt.Errorf("delete by id lock: %w", err)
	}

	defer s.TryUnlock()

	res := s.conn.ZRange(s.ctx, DefaultPrefix+id, 0, -1)

	err := res.Err()
	if err != nil {
		return err
	}

	var errs []error

	groups := res.Val()

	for _, group := range groups {
		errs = append(errs, s.Delete(group, id))
	}

	return errors.Join(errs...)
}

func (s *Redis) GetKeys(group string) []string {
	s.TryLock()
	defer s.TryUnlock()

	res := s.conn.ZRange(s.ctx, DefaultPrefix+group, 0, -1)
	keys := res.Val()

	err := res.Err()
	if err != nil {
		slog.Default().Error("Error get list of keys", "func", "GetKeys", "err", err)
	}

	return keys
}

func (s *Redis) GetIDs() []string {
	s.TryLock()
	defer s.TryUnlock()

	res := s.conn.ZRange(s.ctx, DefaultPrefix+s.key, 0, -1)
	keys := res.Val()

	err := res.Err()
	if err != nil {
		slog.Default().Error("Error get all ids", "func", "GetIDs", "err", err)
	}

	return keys
}

func (s *Redis) Size(group string) int {
	s.TryLock()
	defer s.TryUnlock()

	res := s.conn.ZCount(s.ctx, DefaultPrefix+group, "-inf", "+inf")

	err := res.Err()
	if err != nil {
		slog.Default().Error("Error get size from redis", "err", err)
	}

	return int(res.Val())
}
