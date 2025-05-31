package hll

import "github.com/aerospike/aerospike-client-go/v7"

type Operator interface {
	Operate(*aerospike.WritePolicy, *aerospike.Key, ...*aerospike.Operation) (*aerospike.Record, aerospike.Error)
}
