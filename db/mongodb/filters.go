package mongodb

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	TwoPrecision = 2

	sortAscending  = 1
	sortDescending = -1
)

// ErrFilterPairCount reports an odd key/value filter input.
var ErrFilterPairCount = errors.New("filter pairs must contain key and value")

// ErrFilterKeyType reports a non-strings filter key.
var ErrFilterKeyType = errors.New("filter key must be a strings")

// Field is a core-owned document field.
type Field struct {
	Name  string
	Value any
}

// Filter is a core-owned MongoDB filter document.
type Filter map[string]any

// Document is a core-owned MongoDB document shape.
type Document map[string]any

// SortField is a core-owned MongoDB sort field.
type SortField struct {
	Name       string
	Descending bool
}

// NewFilter creates a core-owned MongoDB filter document.
func NewFilter(fields ...Field) Filter {
	filter := make(Filter, len(fields))
	for _, field := range fields {
		filter[field.Name] = field.Value
	}

	return filter
}

// NewDocument creates a core-owned MongoDB document.
func NewDocument(fields ...Field) Document {
	document := make(Document, len(fields))
	for _, field := range fields {
		document[field.Name] = field.Value
	}

	return document
}

// FilterFromPairs creates a core-owned MongoDB filter from key/value pairs.
func FilterFromPairs(keyValue ...any) (Filter, error) {
	fields, err := fieldsFromPairs(keyValue...)
	if err != nil {
		return nil, err
	}

	return NewFilter(fields...), nil
}

// DocumentFromPairs creates a core-owned MongoDB document from key/value pairs.
func DocumentFromPairs(keyValue ...any) (Document, error) {
	fields, err := fieldsFromPairs(keyValue...)
	if err != nil {
		return nil, err
	}

	return NewDocument(fields...), nil
}

// NewSort creates a MongoDB-compatible sort document without exposing BSON types in the public signature.
func NewSort(fields ...SortField) any {
	sort := make(bson.D, len(fields))
	for i, field := range fields {
		direction := sortAscending
		if field.Descending {
			direction = sortDescending
		}

		sort[i] = bson.E{Key: field.Name, Value: direction}
	}

	return sort
}

// GetBsonD return bson.D object based on values
//
// Deprecated: use NewFilter, NewDocument, FilterFromPairs, DocumentFromPairs, or NewSort.
func GetBsonD(keyValue ...interface{}) bson.D {
	l := len(keyValue)
	if l == 0 || l%TwoPrecision != 0 {
		return bson.D{}
	}

	d := make(bson.D, l/TwoPrecision)

	var k int

	for i := 0; i < len(keyValue); {
		key, val := keyValue[i], keyValue[i+1]
		d[k].Key = key.(string)
		d[k].Value = val

		i += TwoPrecision
		k++
	}

	return d
}

func fieldsFromPairs(keyValue ...any) ([]Field, error) {
	length := len(keyValue)
	if length == 0 {
		return nil, nil
	}

	if length%TwoPrecision != 0 {
		return nil, ErrFilterPairCount
	}

	fields := make([]Field, 0, length/TwoPrecision)
	for i := 0; i < length; i += TwoPrecision {
		key, ok := keyValue[i].(string)
		if !ok {
			return nil, ErrFilterKeyType
		}

		fields = append(fields, Field{Name: key, Value: keyValue[i+1]})
	}

	return fields, nil
}
