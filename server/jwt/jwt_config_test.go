//go:build unit
// +build unit

package jwt

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestJWTConfig_filepath(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY", "test-data/test-jwt.key")
	t.Setenv("JWT_PUBLIC_KEY", "test-data/test-jwt.pem")

	jwtConfig, err := GetConfigFromEnv()
	testutils.Equal(t, err, nil)

	private, err := jwtConfig.GetPrivateKey()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, len(private) != 0, true)

	public, err := jwtConfig.GetPublicKey()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, len(public) != 0, true)

	testutils.Equal(t, string(private)[:5], "-----")
	testutils.Equal(t, string(public)[:5], "-----")
}

func TestJWTConfig_cert(t *testing.T) {
	t.Setenv("JWT_PRIVATE_KEY", `-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----
`)
	t.Setenv("JWT_PUBLIC_KEY", `-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----
`)

	jwtConfig, err := GetConfigFromEnv()
	testutils.Equal(t, err, nil)

	private, err := jwtConfig.GetPrivateKey()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, len(private) != 0, true)

	public, err := jwtConfig.GetPublicKey()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, len(public) != 0, true)

	testutils.Equal(t, string(private)[:5], "-----")
	testutils.Equal(t, string(public)[:5], "-----")
}
