package gremlin

import (
	"fmt"
	"strings"
	"sync"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type Operation interface {
	Execute(cache *Cache, source *gremlingo.GraphTraversalSource) error
	Result() []*gremlingo.Result
}

func PrepareProperties(properties ...interface{}) map[interface{}]interface{} {
	l := len(properties)
	if l == 0 || l%2 != 0 {
		return map[interface{}]interface{}{}
	}

	var k int

	vProperties := map[interface{}]interface{}{}

	for i := 0; i < len(properties); {
		key, val := properties[i], properties[i+1]
		vProperties[key] = val

		i += 2
		k++
	}

	return vProperties
}

type ResultOp struct {
	result []*gremlingo.Result
	mu     sync.RWMutex
}

func newResultOp() *ResultOp {
	return &ResultOp{}
}

func (o *ResultOp) setResult(res []*gremlingo.Result) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.result = make([]*gremlingo.Result, len(res))
	copy(o.result, res)
}

func (o *ResultOp) Result() []*gremlingo.Result {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.result
}

type UpsertVertexOp struct {
	*ResultOp
	properties map[interface{}]interface{}
	id         string
	label      string
}

func NewUpsertVertexOp(label string, id string, properties ...interface{}) *UpsertVertexOp {
	mProperties := PrepareProperties(properties...)

	return &UpsertVertexOp{
		ResultOp:   newResultOp(),
		label:      label,
		id:         id,
		properties: mProperties,
	}
}

func (o *UpsertVertexOp) Execute(cache *Cache, source *gremlingo.GraphTraversalSource) error {
	res, err := MergeV(source.GetGraphTraversal(), o.label, o.id, o.properties).Next()
	if err != nil {
		return fmt.Errorf("UpsertVertexOp Execute failed: %w", err)
	}

	vertex, err := res.GetVertex()
	if err != nil {
		return fmt.Errorf("UpsertVertexOp failed for get vertex: %w", err)
	}

	err = cache.AddVertex(o.label, o.id, vertex)
	if err != nil {
		return fmt.Errorf("UpsertVertexOp failed to add vertex to cache: %w", err)
	}

	o.setResult([]*gremlingo.Result{res})

	return nil
}

type UpsertEdgeOp struct {
	from VertexGetter
	to   VertexGetter
	*ResultOp
	properties map[interface{}]interface{}
	edge       string
	id         string
}

func NewUpsertEdgeOp(
	edge string,
	id string,
	from VertexGetter,
	to VertexGetter,
	properties ...interface{},
) *UpsertEdgeOp {
	mProperties := PrepareProperties(properties...)

	return &UpsertEdgeOp{
		ResultOp:   newResultOp(),
		edge:       edge,
		id:         id,
		from:       from,
		to:         to,
		properties: mProperties,
	}
}

func (o *UpsertEdgeOp) Execute(cache *Cache, source *gremlingo.GraphTraversalSource) error {
	_, vertex1, err := o.from.Get(cache, source)
	if err != nil {
		return fmt.Errorf("UpsertEdgeOp failed for get vertexLabel (from): %w", err)
	}

	_, vertex2, err := o.to.Get(cache, source)
	if err != nil {
		return fmt.Errorf("UpsertEdgeOp failed for get vertexLabel (to): %w", err)
	}

	res, err := MergeE(source.GetGraphTraversal(), o.edge, o.id, vertex1, vertex2, o.properties).Next()
	if err != nil {
		return fmt.Errorf("UpsertEdgeOp Execute failed: %w", err)
	}

	_, err = res.GetEdge()
	if err != nil {
		return fmt.Errorf("UpsertEdgeOp failed for get edge: %w", err)
	}

	o.setResult([]*gremlingo.Result{res})

	return nil
}

type CallbackOp struct {
	*ResultOp
	fn func(cache *Cache, source *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error)
}

func NewCallbackOp(
	fn func(cache *Cache, source *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error),
) *CallbackOp {
	return &CallbackOp{
		ResultOp: newResultOp(),
		fn:       fn,
	}
}

func (o *CallbackOp) Execute(cache *Cache, source *gremlingo.GraphTraversalSource) error {
	results, err := o.fn(cache, source)
	if err != nil {
		return fmt.Errorf("CallbackOp Execute failed: %w", err)
	}

	o.setResult(results)

	return nil
}

type DropVertexOp struct {
	*ResultOp
	vertex VertexGetter
}

func NewDropVertexOp(vertex VertexGetter) *DropVertexOp {
	return &DropVertexOp{
		ResultOp: newResultOp(),
		vertex:   vertex,
	}
}

func (o *DropVertexOp) Execute(cache *Cache, source *gremlingo.GraphTraversalSource) error {
	id1, vertex1, err := o.vertex.Get(cache, source)
	if err != nil {
		return fmt.Errorf("DropVertexOp failed for get vertexLabel (vertexLabel): %w", err)
	}

	t := source.V(vertex1).OutE().Drop()

	res1, err := t.Next()
	if err != nil && !strings.Contains(err.Error(), "E0903") {
		return fmt.Errorf("DropVertexOp Execute failed: %w", err)
	}

	t = source.V(vertex1).InE().Drop()

	res2, err := t.Next()
	if err != nil && !strings.Contains(err.Error(), "E0903") {
		return fmt.Errorf("DropVertexOp Execute failed: %w", err)
	}

	t = source.V(vertex1).Drop()

	res3, err := t.Next()
	if err != nil && !strings.Contains(err.Error(), "E0903") {
		return fmt.Errorf("DropVertexOp Execute failed: %w", err)
	}

	err = cache.DeleteVertex(vertex1.Label, id1)
	if err != nil {
		return fmt.Errorf("DropVertexOp failed to remove vertex: %w", err)
	}

	o.setResult([]*gremlingo.Result{res1, res2, res3})

	return nil
}
