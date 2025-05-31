//go:generate mockgen -source=operations.go -destination=mocks/models.go
package entity

import (
	"github.com/InsideGallery/core/db/aerospike"
	as "github.com/aerospike/aerospike-client-go/v7"
)

type Operations interface {
	Execute([]*as.Operation) error
	GetNamespace() aerospike.Namespace
	Get(bins ...string) (*as.Record, error)
	GetBin(binName string) (interface{}, error)
	Exists() (bool, as.Error)
}
