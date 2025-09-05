package jwt

import (
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
)

const EnvPrefix = "JWT"

type Config struct {
	PrivateKey string `env:"_PRIVATE_KEY,required,expand" envDefault:"${PWD}/testdata/test-jwt.key"`
	PublicKey  string `env:"_PUBLIC_KEY,required,expand" envDefault:"${PWD}/testdata/test-jwt.pem"`
}

func GetConfigFromEnv() (*Config, error) {
	c := new(Config)

	err := env.ParseWithOptions(c, env.Options{
		Prefix: strings.ToUpper(EnvPrefix),
	})
	if err != nil {
		return nil, err
	}

	return c, err
}

func (c Config) GetPrivateKey() ([]byte, error) {
	if !strings.HasPrefix(c.PrivateKey, "-----") {
		data, err := os.ReadFile(c.PrivateKey)
		if err != nil {
			return nil, err
		}

		c.PrivateKey = string(data)
	}

	return []byte(c.PrivateKey), nil
}

func (c Config) GetPublicKey() ([]byte, error) {
	if !strings.HasPrefix(c.PublicKey, "-----") {
		data, err := os.ReadFile(c.PublicKey)
		if err != nil {
			return nil, err
		}

		c.PublicKey = string(data)
	}

	return []byte(c.PublicKey), nil
}
