// Package gremlin provides Gremlin client, graph operation, and traversal helpers.
//
// New code should construct clients from core-owned options and depend on graph
// contracts at the consumer boundary:
//
//	import "github.com/InsideGallery/core/db/gremlin"
//
//	client, err := gremlin.NewClient(gremlin.Options{URL: "ws://127.0.0.1:8182/gremlin"})
//
// Prefer VertexStore or GraphStore with UpsertVertexOptions, UpsertEdgeOptions,
// CountVerticesOptions, ListValuesOptions, and their result types instead of
// exposing Gremlin SDK traversal values through application interfaces.
//
// Compatibility: legacy traversal helpers and package-level syntax state remain
// available for existing consumers. Prefer explicit options and SyntaxState in
// new code.
package gremlin

import (
	"context"
	"sort"

	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"

	coreerrors "github.com/InsideGallery/core/errors"
)

const propertyKeyValuePairSize = 2

// Options is the core-owned input for creating a Gremlin client.
type Options struct {
	URL string
}

// UpsertVertexOptions is the core-owned input for a vertex upsert.
type UpsertVertexOptions struct {
	Label      string
	ID         string
	Properties map[string]any
}

// VertexRef identifies a vertex without exposing Gremlin SDK vertex values.
type VertexRef struct {
	Label string
	ID    string
}

// UpsertEdgeOptions is the core-owned input for an edge upsert.
type UpsertEdgeOptions struct {
	Label      string
	ID         string
	From       VertexRef
	To         VertexRef
	Properties map[string]any
}

// Comparison identifies a core-owned property comparison operation.
type Comparison string

const (
	// ComparisonEqual applies an equality property filter.
	ComparisonEqual Comparison = "eq"
	// ComparisonGreaterThan applies a greater-than property filter.
	ComparisonGreaterThan Comparison = "gt"
	// ComparisonGreaterThanOrEqual applies a greater-than-or-equal property filter.
	ComparisonGreaterThanOrEqual Comparison = "gte"
	// ComparisonLessThan applies a less-than property filter.
	ComparisonLessThan Comparison = "lt"
	// ComparisonLessThanOrEqual applies a less-than-or-equal property filter.
	ComparisonLessThanOrEqual Comparison = "lte"
)

// PropertyFilter is a core-owned Gremlin property filter.
type PropertyFilter struct {
	Name       string
	Comparison Comparison
	Value      any
}

// CountVerticesOptions is the core-owned input for vertex count queries.
type CountVerticesOptions struct {
	Label   string
	ID      string
	Filters []PropertyFilter
}

// ListValuesOptions is the core-owned input for listing property values.
type ListValuesOptions struct {
	Label    string
	ID       string
	Property string
	Filters  []PropertyFilter
}

// GraphResult is the core-owned result for Gremlin graph operations.
type GraphResult struct {
	Affected int64
}

// CountResult is the core-owned result for count queries.
type CountResult struct {
	Count int64
}

// ValueListResult is the core-owned result for value list queries.
type ValueListResult struct {
	Values []any
}

// VertexStore is the core-owned Gremlin contract for new consumers.
type VertexStore interface {
	UpsertVertex(ctx context.Context, options UpsertVertexOptions) (GraphResult, error)
	CloseGraph(ctx context.Context) error
}

// GraphStore is the core-owned Gremlin contract for vertex, edge, and traversal helper operations.
type GraphStore interface {
	VertexStore
	UpsertEdge(ctx context.Context, options UpsertEdgeOptions) (GraphResult, error)
	CountVertices(ctx context.Context, options CountVerticesOptions) (CountResult, error)
	ListValues(ctx context.Context, options ListValuesOptions) (ValueListResult, error)
}

// NewClient creates a Gremlin client from core-owned options.
func NewClient(options Options) (*Client, error) {
	return New(&ConnectionConfig{URL: options.URL})
}

// UpsertVertex upserts one vertex with core-owned options.
func (c *Client) UpsertVertex(ctx context.Context, options UpsertVertexOptions) (GraphResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return GraphResult{}, coreerrors.WrapBoundary("gremlin", "upsert vertex", err)
	}

	op := NewUpsertVertexOp(options.Label, options.ID, propertiesKeyValues(options.Properties)...)
	if err := c.Execute(NewCache(), op); err != nil {
		return GraphResult{}, coreerrors.WrapBoundary("gremlin", "upsert vertex", err)
	}

	return GraphResult{Affected: int64(len(op.Result()))}, nil
}

// UpsertEdge upserts one edge with core-owned options.
func (c *Client) UpsertEdge(ctx context.Context, options UpsertEdgeOptions) (GraphResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return GraphResult{}, coreerrors.WrapBoundary("gremlin", "upsert edge", err)
	}

	op := NewUpsertEdgeOp(
		options.Label,
		options.ID,
		NewLabelVertexGetter(options.From.Label, options.From.ID),
		NewLabelVertexGetter(options.To.Label, options.To.ID),
		propertiesKeyValues(options.Properties)...,
	)
	if err := c.Execute(NewCache(), op); err != nil {
		return GraphResult{}, coreerrors.WrapBoundary("gremlin", "upsert edge", err)
	}

	return GraphResult{Affected: int64(len(op.Result()))}, nil
}

// CountVertices counts matching vertices without exposing traversal helper types.
func (c *Client) CountVertices(ctx context.Context, options CountVerticesOptions) (CountResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return CountResult{}, coreerrors.WrapBoundary("gremlin", "count vertices", err)
	}

	count, err := WrapCount(applyTraversalOptions(c.S().V(), options.Label, options.ID, options.Filters))
	if err != nil {
		return CountResult{}, coreerrors.WrapBoundary("gremlin", "count vertices", err)
	}

	return CountResult{Count: count}, nil
}

// ListValues returns matching property values without exposing traversal helper types.
func (c *Client) ListValues(ctx context.Context, options ListValuesOptions) (ValueListResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return ValueListResult{}, coreerrors.WrapBoundary("gremlin", "list values", err)
	}

	results, err := WrapValuesToList(
		applyTraversalOptions(c.S().V(), options.Label, options.ID, options.Filters),
		options.Property,
	)
	if err != nil {
		return ValueListResult{}, coreerrors.WrapBoundary("gremlin", "list values", err)
	}

	values := make([]any, 0, len(results))
	for _, result := range results {
		values = append(values, result.GetInterface())
	}

	return ValueListResult{Values: values}, nil
}

// CloseGraph closes the Gremlin client through the core-owned contract.
func (c *Client) CloseGraph(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return coreerrors.WrapBoundary("gremlin", "close", err)
	}

	c.Close()

	return nil
}

func propertiesKeyValues(properties map[string]any) []any {
	if len(properties) == 0 {
		return nil
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	values := make([]any, 0, len(properties)*propertyKeyValuePairSize)
	for _, key := range keys {
		values = append(values, key, properties[key])
	}

	return values
}

func applyTraversalOptions(
	traversal *gremlingo.GraphTraversal,
	label string,
	id string,
	filters []PropertyFilter,
) *gremlingo.GraphTraversal {
	if label != "" {
		traversal = WrapLabelFilter(traversal, label)
	}

	if id != "" {
		traversal = WrapperHasID(traversal, id)
	}

	for _, filter := range filters {
		traversal = applyPropertyFilter(traversal, filter)
	}

	return traversal
}

func applyPropertyFilter(traversal *gremlingo.GraphTraversal, filter PropertyFilter) *gremlingo.GraphTraversal {
	switch filter.Comparison {
	case ComparisonGreaterThan:
		return WrapOrderGt(traversal, filter.Name, filter.Value)
	case ComparisonGreaterThanOrEqual:
		return WrapOrderGte(traversal, filter.Name, filter.Value)
	case ComparisonLessThan:
		return WrapOrderLt(traversal, filter.Name, filter.Value)
	case ComparisonLessThanOrEqual:
		return WrapOrderLte(traversal, filter.Name, filter.Value)
	default:
		return traversal.Has(filter.Name, filter.Value)
	}
}
