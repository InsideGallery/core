package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// ProofOfWork represents the structure of our PoW
type ProofOfWork struct {
	Difficulty int
}

// NewProofOfWork creates a new ProofOfWork with a given difficulty
func NewProofOfWork(difficulty int) *ProofOfWork {
	return &ProofOfWork{Difficulty: difficulty}
}

// Validate checks if a given nonce solves the challenge for a message
func (pow *ProofOfWork) Validate(message string, nonce int) bool {
	hash := sha256.Sum256([]byte(message + strconv.Itoa(nonce)))
	hashString := hex.EncodeToString(hash[:])

	return strings.HasPrefix(hashString, strings.Repeat("0", pow.Difficulty))
}

// FindNonce finds a nonce that solves the PoW challenge
func (pow *ProofOfWork) FindNonce(message string) (int, string) {
	nonce := 0

	for {
		if pow.Validate(message, nonce) {
			hash := sha256.Sum256([]byte(message + strconv.Itoa(nonce)))
			return nonce, hex.EncodeToString(hash[:])
		}

		nonce++
	}
}
