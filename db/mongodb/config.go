package mongodb

import (
	"strings"

	"github.com/caarlos0/env/v10"
)

// EnvPrefix environment prefix for mongodb config
const EnvPrefix = "MONGO"

// ConnectionConfig contains required data for mongo
type ConnectionConfig struct {
	Hosts         []string `env:"_HOSTS" envDefault:""`
	Host          string   `env:"_HOST" envDefault:"localhost"`
	Port          string   `env:"_PORT" envDefault:"27017"`
	User          string   `env:"_USER" envDefault:""`
	Pass          string   `env:"_PASS" envDefault:""`
	Scheme        string   `env:"_SCHEME" envDefault:"mongodb"`
	Database      string   `env:"_DATABASE" envDefault:"default"`
	Args          string   `env:"_ARGS" envDefault:""`
	Mode          string   `env:"_MODE" envDefault:"secondarypreferred"`
	RetryWrites   bool     `env:"_RETRYWRITES" envDefault:"false"`
	AuthMechanism string   `env:"_AUTH_MECHANISM" envDefault:"SCRAM-SHA-256"`
	AuthSource    string   `env:"_AUTH_SOURCE" envDefault:""`
}

// GetConnectionConfigFromEnv return mongodb configs bases on environment variables
func GetConnectionConfigFromEnv() (*ConnectionConfig, error) {
	c := new(ConnectionConfig)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: strings.ToUpper(EnvPrefix),
	})
	if err != nil {
		return nil, err
	}

	return c, err
}

func (c *ConnectionConfig) GetDSN() string {
	dsn := c.Scheme + "://"

	if len(c.Hosts) > 0 {
		dsn = strings.Join([]string{dsn, strings.Join(c.Hosts, ",")}, "")
	} else {
		dsn = strings.Join([]string{dsn, c.Host, ":", c.Port}, "")
	}

	if c.Args != "" {
		dsn = strings.Join([]string{dsn, "/?", c.Args}, "")
	}

	return dsn
}
