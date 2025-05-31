package neo4j

import (
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/auth"
)

// EnvPrefix environment prefix for gremlin config
const EnvPrefix = "NEO4J"

const (
	TypeBasicAuth    = "BasicAuth"
	TypeKerberosAuth = "KerberosAuth"
	TypeBearerAuth   = "BearerAuth"
)

// ConnectionConfig contains required data for gremlin
type ConnectionConfig struct {
	Login    string `env:"_LOGIN" envDefault:"neo4j"`
	Password string `env:"_PASSWORD" envDefault:""`
	Realm    string `env:"_REALM" envDefault:""`
	Ticket   string `env:"_TICKET" envDefault:""`
	Token    string `env:"_TOKEN" envDefault:""`
	Host     string `env:"_HOST" envDefault:"neo4j://127.0.0.1:8080"`
	TypeAuth string `env:"_AUTH" envDefault:"NoAuth"`
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

func (c *ConnectionConfig) TokenManager(m auth.TokenManager) auth.TokenManager {
	if m != nil {
		return m
	}

	switch c.TypeAuth {
	case TypeBasicAuth:
		return neo4j.BasicAuth(c.Login, c.Password, c.Realm)
	case TypeKerberosAuth:
		return neo4j.KerberosAuth(c.Ticket)
	case TypeBearerAuth:
		return neo4j.BearerAuth(c.Token)
	}

	return neo4j.NoAuth()
}
