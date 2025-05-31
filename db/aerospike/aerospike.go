package aerospike

import (
	"time"

	aero "github.com/aerospike/aerospike-client-go/v7"
)

type Mapper interface {
	ToMap() map[string]interface{}
}

type SetKey struct {
	Key interface{}
	Set string
}

func NewValue(value interface{}) aero.Value {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case Mapper:
		return aero.NewValue(v.ToMap())
	case *uint64:
		return aero.NewLongValue(int64(*v))
	case *uint32:
		return aero.NewIntegerValue(int(*v))
	case *time.Time:
		return aero.NewLongValue(v.Unix())
	case *bool:
		return aero.NewValue(*v)
	case *int64:
		return aero.NewValue(*v)
	case *aero.GeoJSONValue:
		return *v
	case aero.GeoJSONValue:
		return v
	case uint64:
		return aero.NewLongValue(int64(v))
	case uint32:
		return aero.NewIntegerValue(int(v))
	case time.Time:
		return aero.NewLongValue(v.Unix())
	}

	return aero.NewValue(value)
}

func NewBin(name string, value interface{}) *aero.Bin {
	return aero.NewBin(name, NewValue(value))
}

type NamespaceInstance struct {
	conn      *aero.Client
	namespace string
}

func NewNamespaceInstance(namespace string, names ...string) (*NamespaceInstance, error) {
	conn, err := Default(names...)
	if err != nil {
		return nil, err
	}

	return &NamespaceInstance{
		namespace: namespace,
		conn:      conn,
	}, nil
}

func (ni *NamespaceInstance) Put(
	policy *aero.WritePolicy,
	setName string,
	value interface{},
	bins aero.BinMap,
) aero.Error {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return err
	}

	return ni.conn.Put(policy, key, bins)
}

func (ni *NamespaceInstance) PutBins(
	policy *aero.WritePolicy,
	setName string,
	value interface{},
	bins ...*aero.Bin,
) aero.Error {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return err
	}

	return ni.conn.PutBins(policy, key, bins...)
}

func (ni *NamespaceInstance) Query(policy *aero.QueryPolicy, stmt *aero.Statement) (*aero.Recordset, aero.Error) {
	return ni.conn.Query(policy, stmt)
}

func (ni *NamespaceInstance) CreateIndex(
	policy *aero.WritePolicy,
	setName string,
	indexName string,
	binName string,
	indexType aero.IndexType,
) (*aero.IndexTask, aero.Error) {
	return ni.conn.CreateIndex(policy, ni.namespace, setName, indexName, binName, indexType)
}

func (ni *NamespaceInstance) DropIndex(
	policy *aero.WritePolicy,
	setName string,
	indexName string,
) aero.Error {
	return ni.conn.DropIndex(policy, ni.namespace, setName, indexName)
}

func (ni *NamespaceInstance) CreateComplexIndex(
	policy *aero.WritePolicy,
	setName string,
	indexName string,
	binName string,
	indexType aero.IndexType,
	indexCollectionType aero.IndexCollectionType,
	ctx ...*aero.CDTContext,
) (*aero.IndexTask, aero.Error) {
	return ni.conn.
		CreateComplexIndex(policy, ni.namespace, setName, indexName, binName, indexType, indexCollectionType, ctx...)
}

func (ni *NamespaceInstance) Get(
	policy *aero.BasePolicy,
	setName string,
	value interface{},
	binNames ...string,
) (*aero.Record, aero.Error) {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return nil, err
	}

	return ni.conn.Get(policy, key, binNames...)
}

func (ni *NamespaceInstance) Operate(
	policy *aero.WritePolicy,
	setName string,
	value interface{},
	ops ...*aero.Operation,
) (*aero.Record, aero.Error) {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return nil, err
	}

	return ni.conn.Operate(policy, key, ops...)
}

func (ni *NamespaceInstance) Delete(policy *aero.WritePolicy, setName string, value interface{}) (bool, aero.Error) {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return false, err
	}

	return ni.conn.Delete(policy, key)
}

func (ni *NamespaceInstance) BatchDelete(
	policy *aero.BatchPolicy,
	deletePolicy *aero.BatchDeletePolicy,
	values []SetKey,
) ([]*aero.BatchRecord, aero.Error) {
	keys := make([]*aero.Key, 0, len(values))

	for _, v := range values {
		key, err := aero.NewKey(ni.namespace, v.Set, v.Key)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return ni.conn.BatchDelete(policy, deletePolicy, keys)
}

func (ni *NamespaceInstance) BatchGet(
	policy *aero.BatchPolicy,
	values []SetKey,
	names ...string,
) ([]*aero.Record, aero.Error) {
	keys := make([]*aero.Key, 0, len(values))

	for _, v := range values {
		key, err := aero.NewKey(ni.namespace, v.Set, v.Key)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return ni.conn.BatchGet(policy, keys, names...)
}

func (ni *NamespaceInstance) QueryAggregate(
	policy *aero.QueryPolicy,
	statement *aero.Statement,
	packageName, functionName string,
	functionArgs ...aero.Value,
) (*aero.Recordset, aero.Error) {
	return ni.conn.QueryAggregate(policy, statement, packageName, functionName, functionArgs...)
}

func (ni *NamespaceInstance) Exists(policy *aero.BasePolicy, setName string, value interface{}) (bool, aero.Error) {
	key, err := aero.NewKey(ni.namespace, setName, value)
	if err != nil {
		return false, err
	}

	return ni.conn.Exists(policy, key)
}

func (ni *NamespaceInstance) Truncate(policy *aero.InfoPolicy, set string, beforeLastUpdate *time.Time) aero.Error {
	return ni.conn.Truncate(policy, ni.namespace, set, beforeLastUpdate)
}

func (ni *NamespaceInstance) Stats() (map[string]interface{}, aero.Error) {
	return ni.conn.Stats()
}

func (ni *NamespaceInstance) GetNamespace() string {
	return ni.namespace
}

func (ni *NamespaceInstance) GetConnection() *aero.Client {
	return ni.conn
}

func (ni *NamespaceInstance) BatchOperate(policy *aero.BatchPolicy, records []aero.BatchRecordIfc) aero.Error {
	return ni.conn.BatchOperate(policy, records)
}
