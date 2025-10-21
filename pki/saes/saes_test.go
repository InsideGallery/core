package saes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func TestSAESCipher(t *testing.T) {
	a, err := NewSAES()
	require.NoError(t, err)
	require.NotNil(t, a)

	val := []byte("test string")
	res, err := a.Encrypt(val)
	assert.NoError(t, err)

	original, err := a.Decrypt(res)
	assert.NoError(t, err)

	assert.Equal(t, val, original)
}

func TestStaticAESCipher(t *testing.T) {
	a, err := NewSAES()
	require.NoError(t, err)
	require.NotNil(t, a)

	val := []byte("test string")
	res, err := a.Encrypt(val)
	assert.NoError(t, err)

	res2, err := a.Encrypt(val)
	assert.NoError(t, err)

	assert.Equal(t, res, res2)

	original, err := a.Decrypt(res)
	assert.NoError(t, err)

	assert.Equal(t, val, original)
}

func TestAESCipherEncryptedDataIsEmpty(t *testing.T) {
	a, err := NewSAES()
	require.NoError(t, err)
	require.NotNil(t, a)

	var expected []byte

	original, err := a.Decrypt(nil)
	assert.ErrorIs(t, err, ErrEncryptedDataIsEmpty)

	assert.Equal(t, original, expected)
}

func TestAESCipherLong(t *testing.T) {
	a, err := NewSAES()
	require.NoError(t, err)
	require.NotNil(t, a)

	val := bytes.Repeat([]byte{127}, 10000)
	res, err := a.Encrypt(val)
	assert.NoError(t, err)

	original, err := a.Decrypt(res)
	assert.NoError(t, err)

	assert.Equal(t, val, original)
}

func TestAESCipherRestore(t *testing.T) {
	a, err := NewSAES()
	require.NoError(t, err)
	require.NotNil(t, a)

	raw, err := a.ToBinary()
	assert.NoError(t, err)

	c, err := a.FromBinary(raw)
	assert.NoError(t, err)

	val := []byte("test string")
	res, err := c.Encrypt(val)
	assert.NoError(t, err)

	original, err := c.Decrypt(res)
	assert.NoError(t, err)

	assert.Equal(t, val, original)
}

func BenchmarkNewSAESEncrypt10(b *testing.B) {
	c, err := NewSAES()
	if err != nil {
		b.Fatalf("failed to init random SAES cipher: %s", err.Error())
	}

	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "random_10_bytes",
			input: random.GetRandomBytes(10),
		},
		{
			name:  "random_100_bytes",
			input: random.GetRandomBytes(100),
		},
		{
			name:  "random_1000_bytes",
			input: random.GetRandomBytes(1000),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_, err = c.Encrypt(tt.input)
				if err != nil {
					b.Fatalf("encryption failed: %s", err.Error())
				}
			}
		})
	}
}
