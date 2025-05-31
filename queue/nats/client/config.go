package client

import (
	"errors"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

const (
	DefaultEnvNATSPrefix = "NATS"

	DefaultTimeout = time.Millisecond * 100
)

type Config struct {
	*nats.StreamConfig
	Addr                 string        `env:"_ADDR" envDefault:"nats://127.0.0.1:4222"`
	Username             string        `env:"_USERNAME" envDefault:""`
	Password             string        `env:"_PASSWORD" envDefault:""`
	Seed                 string        `env:"_SEED" envDefault:""`
	DrainTimeout         time.Duration `env:"_DRAIN_TIMEOUT" envDefault:"1s"`
	MaxReconnects        int           `env:"_MAX_RECONNECTS" envDefault:"10"`
	ReconnectWait        time.Duration `env:"_RECONNECT_WAIT" envDefault:"1s"`
	MaxAckPending        int           `env:"_MAX_ACK_PENDING" envDefault:"0"`
	RetryOnFailedConnect bool          `env:"_RETRY_ON_FAILED_CONNECT" envDefault:"true"`
	ManualAck            bool          `env:"_MANUAL_ACK" envDefault:"false"`
	ConcurrentSize       int           `env:"_CONCURRENT_SIZE" envDefault:"10"`
	MaxConcurrentSize    uint64        `env:"_MAX_CONCURRENT_SIZE" envDefault:"1024"`
	ReadTimeout          time.Duration `env:"_READ_TIMEOUT" envDefault:"500ms"`
	IdleTimeout          time.Duration `env:"_IDLE_TIMEOUT" envDefault:"5s"`
}

func GetNATSConnectionConfigFromEnv(prefixes ...string) (*Config, error) {
	c := new(Config)

	prefix := DefaultEnvNATSPrefix
	if len(prefixes) > 0 {
		prefix = prefixes[0]
	}

	err := env.ParseWithOptions(c, env.Options{
		Prefix: strings.ToUpper(prefix),
	})
	if err != nil {
		return nil, err
	}

	return c, err
}

func (cfg *Config) GetConcurrentSize() int {
	if cfg.ConcurrentSize <= 0 {
		return runtime.NumCPU()
	}

	return cfg.ConcurrentSize
}

func (cfg *Config) GetReadTimeout() time.Duration {
	if cfg.ReadTimeout <= 0 {
		return DefaultTimeout
	}

	return cfg.ReadTimeout
}

func (cfg *Config) GetOptions() []nats.Option {
	options := []nats.Option{
		nats.RetryOnFailedConnect(cfg.RetryOnFailedConnect),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	if cfg.Username != "" && cfg.Password != "" {
		options = append(options, nats.UserInfo(cfg.Username, cfg.Password))
	}

	if cfg.DrainTimeout > 0 {
		options = append(options, nats.DrainTimeout(cfg.DrainTimeout))
	}

	if cfg.Seed != "" {
		kp, err := nkeys.FromSeed([]byte(cfg.Seed))
		if err != nil {
			slog.Default().Error("Error getting key from seed", "err", err)

			return options
		}

		usrNKey, err := kp.PublicKey()
		if err != nil {
			slog.Default().Error("Error getting public key from key", "err", err)

			return options
		}

		options = append(options, nats.Nkey(usrNKey, func(nonce []byte) ([]byte, error) {
			return kp.Sign(nonce)
		}))
	}

	options = append(options, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		cid, cerr := nc.GetClientID()
		if cerr != nil {
			err = errors.Join(cerr, err)
		}

		if sub != nil {
			slog.Error("Error on connection",
				"err", err, "cid", cid, "subject", sub.Subject)
		} else {
			slog.Error("Error on connection",
				"err", err, "cid", cid)
		}
	}))

	return options
}
