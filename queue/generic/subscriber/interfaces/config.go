//go:generate mockgen -package mock -source=config.go -destination=mock/config.go
package interfaces

import (
	"time"
)

type Config interface {
	GetReadTimeout() time.Duration
	GetMaxConcurrentSize() uint64
	GetConcurrentSize() int
}
