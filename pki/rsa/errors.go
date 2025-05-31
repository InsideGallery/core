package rsa

import "errors"

// All kind of errors
var (
	ErrFailedToParsePEMBlock = errors.New("failed to parse PEM block containing the public key")
)
