//go:generate mockgen -package mock -source=interface.go -destination=mock/proxy.go
package proxy

import (
	"context"
	"time"
)

type NATSPopulator interface {
	NATSPublisher
	NATSRequester
}

type NATSPublisher interface {
	Publish(subj string, msg []byte, headers ...string) error
	PublishWithContext(ctx context.Context, subj string, msg []byte, headers ...string) error
}

type NATSRequester interface {
	Requester(subj string, data []byte, timeout time.Duration, headers ...string) ([]byte, error)
	RequesterWithContext(ctx context.Context, subj string, msg []byte, headers ...string) ([]byte, error)
}

type Storage interface {
	Add(group string, id string) error
	Delete(group string, id string) error
	DeleteByID(id string) error
	GetKeys(group string) []string
	GetIDs() []string
	Size(group string) int
}
