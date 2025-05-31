package ecs

import (
	"context"
)

// System is an interface which implements an ECS-System
type System interface {
	Update(ctx context.Context) error
}
