package ecs

import (
	"context"
)

// Component is an interface which implements an ECS-System
type Component interface {
	Update(ctx context.Context) error
}
