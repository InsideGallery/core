package gremlin

import (
	"errors"
	"fmt"

	"github.com/InsideGallery/core/memory/registry"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type VertexGetter interface {
	Get(cache *Cache, source *gremlingo.GraphTraversalSource) (string, *gremlingo.Vertex, error)
}

type LabelVertexGetter struct {
	label string
	id    string
}

func NewLabelVertexGetter(label string, id string) LabelVertexGetter {
	return LabelVertexGetter{
		label: label,
		id:    id,
	}
}

func (v LabelVertexGetter) Get(
	cache *Cache,
	source *gremlingo.GraphTraversalSource,
) (string, *gremlingo.Vertex, error) {
	if Syntax == SyntaxNeptun {
		return v.id, &gremlingo.Vertex{
			Element: gremlingo.Element{
				Id:    v.id,
				Label: v.label,
			},
		}, nil
	}

	vertex, err := cache.GetVertex(v.label, v.id)
	if err != nil && !errors.Is(err, registry.ErrNotFoundEntity) {
		return v.id, vertex, err
	}

	if err == nil {
		return v.id, vertex, nil
	}

	res, err := WrapperHasID(source.V(), v.id).HasLabel(v.label).Next()
	if err != nil {
		return v.id, nil, fmt.Errorf("LabelVertexGetter failed to get result: %w", err)
	}

	vertex, err = res.GetVertex()
	if err != nil {
		return v.id, nil, fmt.Errorf("LabelVertexGetter failed to get vertex: %w", err)
	}

	return v.id, vertex, err
}

type CommonVertexGetter struct {
	vertex *gremlingo.Vertex
	id     string
}

func NewCommonVertexGetter(vertex *gremlingo.Vertex, id string) CommonVertexGetter {
	return CommonVertexGetter{
		vertex: vertex,
		id:     id,
	}
}

func (v CommonVertexGetter) Get(_ *Cache, _ *gremlingo.GraphTraversalSource) (string, *gremlingo.Vertex, error) {
	return v.id, v.vertex, nil
}
