//go:generate mockgen -package mock -source=config.go -destination=mock/config.go
package interfaces

import (
	"time"
)

type Config interface {
	// Deprecated: use ReadTimeout.
	GetReadTimeout() time.Duration
	// Deprecated: use MaxConcurrentSize.
	GetMaxConcurrentSize() uint64
	// Deprecated: use ConcurrentSize.
	GetConcurrentSize() int
}

type readTimeoutConfig interface {
	ReadTimeout() time.Duration
}

type maxConcurrentSizeConfig interface {
	MaxConcurrentSize() uint64
}

type concurrentSizeConfig interface {
	ConcurrentSize() int
}

// ReadTimeout returns config read timeout.
func ReadTimeout(config Config) time.Duration {
	if valueConfig, ok := config.(readTimeoutConfig); ok {
		return valueConfig.ReadTimeout()
	}

	return config.GetReadTimeout()
}

// MaxConcurrentSize returns config maximum concurrent size.
func MaxConcurrentSize(config Config) uint64 {
	if valueConfig, ok := config.(maxConcurrentSizeConfig); ok {
		return valueConfig.MaxConcurrentSize()
	}

	return config.GetMaxConcurrentSize()
}

// ConcurrentSize returns config concurrent size.
func ConcurrentSize(config Config) int {
	if valueConfig, ok := config.(concurrentSizeConfig); ok {
		return valueConfig.ConcurrentSize()
	}

	return config.GetConcurrentSize()
}
