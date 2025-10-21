package saes

import "errors"

var (
	ErrEncryptedDataIsEmpty    = errors.New("encrypted data is empty")
	ErrEncryptedDataIsWrongLen = errors.New("encrypted data is wrong len")
)
