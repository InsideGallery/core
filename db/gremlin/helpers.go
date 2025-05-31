package gremlin

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func WrapperHasID(t *gremlingo.GraphTraversal, id string) *gremlingo.GraphTraversal {
	switch Syntax {
	case SyntaxAerospike:
		return t.Has(PropertyID, id)
	case SyntaxNeptun:
		return t.HasId(id)
	}

	return t
}

func WrapLabelFilter(t *gremlingo.GraphTraversal, label string) *gremlingo.GraphTraversal {
	return t.HasLabel(label)
}

func WrapOrderDesc(t *gremlingo.GraphTraversal, property string) *gremlingo.GraphTraversal {
	return t.Order().By(property, gremlingo.Order.Desc)
}

func WrapOrderAsc(t *gremlingo.GraphTraversal, property string) *gremlingo.GraphTraversal {
	return t.Order().By(property, gremlingo.Order.Asc)
}

func WrapOrderGt(t *gremlingo.GraphTraversal, property string, value interface{}) *gremlingo.GraphTraversal {
	return t.Has(property, gremlingo.T__.Is(gremlingo.P.Gt(value)))
}

func WrapOrderLt(t *gremlingo.GraphTraversal, property string, value interface{}) *gremlingo.GraphTraversal {
	return t.Has(property, gremlingo.T__.Is(gremlingo.P.Lt(value)))
}

func WrapOrderGte(t *gremlingo.GraphTraversal, property string, value interface{}) *gremlingo.GraphTraversal {
	return t.Has(property, gremlingo.T__.Is(gremlingo.P.Gte(value)))
}

func WrapOrderLte(t *gremlingo.GraphTraversal, property string, value interface{}) *gremlingo.GraphTraversal {
	return t.Has(property, gremlingo.T__.Is(gremlingo.P.Lte(value)))
}

func WrapValuesToList(t *gremlingo.GraphTraversal, values string) ([]*gremlingo.Result, error) {
	return t.Values(values).ToList()
}

func MergeV(
	t *gremlingo.GraphTraversal,
	label, id string,
	properties map[interface{}]interface{},
) *gremlingo.GraphTraversal {
	createProperties := map[interface{}]interface{}{}
	merge := map[interface{}]interface{}{
		gremlingo.T.Label: label,
		PropertyID:        id,
	}

	switch Syntax {
	case SyntaxNeptun:
		createProperties[gremlingo.T.Id] = id
	case SyntaxAerospike:
		properties[PropertyID] = id
		createProperties[gremlingo.T.Label] = label
	}

	for k, v := range properties {
		createProperties[k] = v
	}

	return t.MergeV(merge).
		Option(gremlingo.Merge.OnMatch, properties).
		Option(gremlingo.Merge.OnCreate, createProperties)
}

func MergeE(
	t *gremlingo.GraphTraversal,
	edge string,
	id string,
	vertex1, vertex2 *gremlingo.Vertex,
	properties map[interface{}]interface{},
) *gremlingo.GraphTraversal {
	createProperties := map[interface{}]interface{}{}
	merge := map[interface{}]interface{}{
		gremlingo.T.Label:        edge,
		gremlingo.Direction.From: vertex1,
		gremlingo.Direction.To:   vertex2,
	}

	switch Syntax {
	case SyntaxNeptun:
		merge[gremlingo.T.Id] = id
	case SyntaxAerospike:
		properties[PropertyID] = id
		createProperties[gremlingo.T.Label] = edge
	}

	for k, v := range properties {
		createProperties[k] = v
	}

	return t.MergeE(merge).
		Option(gremlingo.Merge.OnMatch, properties).
		Option(gremlingo.Merge.OnCreate, createProperties)
}

func WrapCount(t *gremlingo.GraphTraversal) (int64, error) {
	res, err := t.Count().Next()
	if err != nil {
		return 0, err
	}

	return res.GetInt64()
}
