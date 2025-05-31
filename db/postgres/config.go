package postgres

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

// EnvPrefix environment prefix for mongodb config
const EnvPrefix = "POSTGRES"

// ConnectionConfig contains required data for gremlin
type ConnectionConfig struct {
	Host            string `env:"_HOST" envDefault:"localhost"`
	Port            string `env:"_PORT" envDefault:"5432"`
	User            string `env:"_USER" envDefault:"default"`
	Password        string `env:"_PASSWORD" envDefault:"default"`
	DB              string `env:"_DB" envDefault:"default"`
	ApplicationName string `env:"_APPLICATIONNAME" envDefault:""`
	MaxOpenConns    int    `env:"_MAXOPENCONNS" envDefault:"500"`
	ConnMaxLifetime int64  `env:"_CONNMAXLIFETIME" envDefault:"-1"`
}

// GetConnectionConfigFromEnv return aerospike configs bases on environment variables
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
	if c.ApplicationName == "" {
		return fmt.Sprintf(
			"port=%s dbname=%s user=%s password=%s host=%s sslmode=disable",
			c.Port, c.DB, c.User, c.Password, c.Host,
		)
	}

	return fmt.Sprintf(
		"port=%s dbname=%s user=%s password=%s host=%s fallback_application_name=%s sslmode=disable",
		c.Port, c.DB, c.User, c.Password, c.Host, c.ApplicationName,
	)
}
