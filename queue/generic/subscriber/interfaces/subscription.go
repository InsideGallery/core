//go:generate mockgen -package mock -source=subscription.go -destination=mock/subscription.go
package interfaces

import "time"

type Subscription interface {
	NextMsg(timeout time.Duration) (Msg, error)
	Drain() error
	GetSubject() string
	Pending() (int64, int64, error)
	Dropped() (int64, error)
	Delivered() (int64, error)
}
