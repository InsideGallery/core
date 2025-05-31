//go:generate mockgen -source=interface.go -destination=mocks/interface.go
package aerospike

import (
	"time"

	"github.com/aerospike/aerospike-client-go/v7"
)

type Aerospike interface {
	IsConnected() bool
	GetNodeNames() []string
	Put(*aerospike.WritePolicy, *aerospike.Key, aerospike.BinMap) aerospike.Error
	PutBins(*aerospike.WritePolicy, *aerospike.Key, ...*aerospike.Bin) aerospike.Error
	Query(*aerospike.QueryPolicy, *aerospike.Statement) (*aerospike.Recordset, aerospike.Error)

	CreateIndex(*aerospike.WritePolicy,
		string,
		string,
		string,
		string,
		aerospike.IndexType,
	) (*aerospike.IndexTask, aerospike.Error)

	DropIndex(
		policy *aerospike.WritePolicy,
		namespace string,
		setName string,
		indexName string,
	) aerospike.Error

	CreateComplexIndex(
		*aerospike.WritePolicy,
		string,
		string,
		string,
		string,
		aerospike.IndexType,
		aerospike.IndexCollectionType,
		...*aerospike.CDTContext,
	) (*aerospike.IndexTask, aerospike.Error)

	Stats() (map[string]interface{}, aerospike.Error)
	Get(*aerospike.BasePolicy, *aerospike.Key, ...string) (*aerospike.Record, aerospike.Error)
	Operate(*aerospike.WritePolicy, *aerospike.Key, ...*aerospike.Operation) (*aerospike.Record, aerospike.Error)
	Delete(
		policy *aerospike.WritePolicy,
		key *aerospike.Key,
	) (bool, aerospike.Error)
	BatchDelete(
		policy *aerospike.BatchPolicy,
		deletePolicy *aerospike.BatchDeletePolicy,
		keys []*aerospike.Key,
	) ([]*aerospike.BatchRecord, aerospike.Error)
	BatchGet(*aerospike.BatchPolicy, []*aerospike.Key, ...string) ([]*aerospike.Record, aerospike.Error)
	RegisterUDF(
		policy *aerospike.WritePolicy,
		udfBody []byte,
		serverPath string,
		language aerospike.Language,
	) (*aerospike.RegisterTask, aerospike.Error)
	RegisterUDFFromFile(
		policy *aerospike.WritePolicy,
		clientPath string,
		serverPath string,
		language aerospike.Language,
	) (*aerospike.RegisterTask, aerospike.Error)
	ExecuteUDF(
		policy *aerospike.QueryPolicy,
		statement *aerospike.Statement,
		packageName string,
		functionName string,
		functionArgs ...aerospike.Value,
	) (*aerospike.ExecuteTask, aerospike.Error)
	QueryAggregate(
		policy *aerospike.QueryPolicy,
		statement *aerospike.Statement,
		packageName, functionName string,
		functionArgs ...aerospike.Value,
	) (*aerospike.Recordset, aerospike.Error)
	Exists(policy *aerospike.BasePolicy, key *aerospike.Key) (bool, aerospike.Error)

	Truncate(policy *aerospike.InfoPolicy, namespace, set string, beforeLastUpdate *time.Time) aerospike.Error
	BatchOperate(policy *aerospike.BatchPolicy, records []aerospike.BatchRecordIfc) aerospike.Error
}

type Namespace interface {
	Put(*aerospike.WritePolicy, string, interface{}, aerospike.BinMap) aerospike.Error
	PutBins(policy *aerospike.WritePolicy, setName string, value interface{}, bins ...*aerospike.Bin) aerospike.Error
	Query(*aerospike.QueryPolicy, *aerospike.Statement) (*aerospike.Recordset, aerospike.Error)

	CreateIndex(
		policy *aerospike.WritePolicy,
		set string,
		indexName string,
		fieldName string,
		indexType aerospike.IndexType,
	) (*aerospike.IndexTask, aerospike.Error)

	DropIndex(
		policy *aerospike.WritePolicy,
		setName string,
		indexName string,
	) aerospike.Error

	CreateComplexIndex(
		policy *aerospike.WritePolicy,
		set string,
		indexName string,
		field string,
		indexType aerospike.IndexType,
		collectionType aerospike.IndexCollectionType,
		ctx ...*aerospike.CDTContext,
	) (*aerospike.IndexTask, aerospike.Error)

	Stats() (map[string]interface{}, aerospike.Error)
	Get(
		policy *aerospike.BasePolicy,
		setName string,
		value interface{},
		binNames ...string,
	) (*aerospike.Record, aerospike.Error)
	Operate(
		policy *aerospike.WritePolicy,
		setName string,
		value interface{},
		ops ...*aerospike.Operation,
	) (*aerospike.Record, aerospike.Error)
	Delete(
		policy *aerospike.WritePolicy,
		setName string,
		value interface{},
	) (bool, aerospike.Error)
	BatchDelete(
		policy *aerospike.BatchPolicy,
		deletePolicy *aerospike.BatchDeletePolicy,
		values []SetKey,
	) ([]*aerospike.BatchRecord, aerospike.Error)
	BatchGet(*aerospike.BatchPolicy, []SetKey, ...string) ([]*aerospike.Record, aerospike.Error)
	QueryAggregate(
		policy *aerospike.QueryPolicy,
		statement *aerospike.Statement,
		packageName, functionName string,
		functionArgs ...aerospike.Value,
	) (*aerospike.Recordset, aerospike.Error)
	Exists(*aerospike.BasePolicy, string, interface{}) (bool, aerospike.Error)

	Truncate(policy *aerospike.InfoPolicy, set string, beforeLastUpdate *time.Time) aerospike.Error
	GetConnection() Aerospike
	GetNamespace() string
	BatchOperate(policy *aerospike.BatchPolicy, records []aerospike.BatchRecordIfc) aerospike.Error
}
