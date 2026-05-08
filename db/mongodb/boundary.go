// Package mongodb provides MongoDB client and document helpers.
//
// New code should use explicit client ownership and core-owned document
// contracts:
//
//	import "github.com/InsideGallery/core/db/mongodb"
//
//	store := mongodb.NewClientStore(nil)
//	client, err := store.GetOrCreate(config)
//
// Prefer DocumentStore with FindOptions, CountOptions, InsertOptions,
// UpdateOptions, DeleteOptions, DocumentResult, and WriteResult for application
// boundaries that should not expose MongoDB SDK result types.
//
// Compatibility: package-level Set, Get, and Default remain available for
// existing consumers. Prefer NewMongoClient or ClientStore.GetOrCreate with
// explicit configuration in new code.
package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"

	coreerrors "github.com/InsideGallery/core/errors"
)

// ErrDocumentTargetIsNotSet reports a missing decode target for document reads.
var ErrDocumentTargetIsNotSet = errors.New("document target is not set")

// FindOptions is the core-owned input for MongoDB read operations.
type FindOptions struct {
	Collection string
	Filter     any
	Target     any
	Limit      int64
	Skip       int64
	Sort       any
}

// CountOptions is the core-owned input for MongoDB count operations.
type CountOptions struct {
	Collection string
	Filter     any
}

// InsertOptions is the core-owned input for MongoDB insert operations.
type InsertOptions struct {
	Collection string
	Document   any
}

// UpdateOptions is the core-owned input for MongoDB update operations.
type UpdateOptions struct {
	Collection string
	Filter     any
	Update     any
	Upsert     bool
	Many       bool
}

// DeleteOptions is the core-owned input for MongoDB delete operations.
type DeleteOptions struct {
	Collection string
	Filter     any
	Many       bool
}

// DocumentResult is the core-owned result for MongoDB read operations.
type DocumentResult struct {
	Found     bool
	Document  any
	Documents []any
	Count     int64
}

// WriteResult is the core-owned result for MongoDB write operations.
type WriteResult struct {
	InsertedCount int64
	MatchedCount  int64
	ModifiedCount int64
	DeletedCount  int64
	UpsertedCount int64
	InsertedID    any
	UpsertedID    any
}

// DocumentStore is the core-owned MongoDB contract for new consumers.
type DocumentStore interface {
	FindOneDocument(ctx context.Context, options FindOptions) (DocumentResult, error)
	FindDocuments(ctx context.Context, options FindOptions) (DocumentResult, error)
	Count(ctx context.Context, options CountOptions) (DocumentResult, error)
	InsertDocument(ctx context.Context, options InsertOptions) (WriteResult, error)
	UpdateDocuments(ctx context.Context, options UpdateOptions) (WriteResult, error)
	DeleteDocuments(ctx context.Context, options DeleteOptions) (WriteResult, error)
}

// FindOneDocument reads one document with core-owned options.
func (m *MongoClient) FindOneDocument(ctx context.Context, options FindOptions) (DocumentResult, error) {
	if options.Target == nil {
		return DocumentResult{}, ErrDocumentTargetIsNotSet
	}

	err := m.FindOne(ctx, options.Collection, options.Target, options.Filter, options.findOneOptions())
	if err != nil {
		return DocumentResult{}, coreerrors.WrapBoundary("mongodb", "find one document", err)
	}

	return DocumentResult{Found: true, Document: options.Target}, nil
}

// FindDocuments reads documents with core-owned options.
func (m *MongoClient) FindDocuments(ctx context.Context, options FindOptions) (DocumentResult, error) {
	if options.Target == nil {
		return DocumentResult{}, ErrDocumentTargetIsNotSet
	}

	documents, err := m.Find(ctx, options.Collection, options.Target, options.Filter, options.findOptions())
	if err != nil {
		return DocumentResult{}, coreerrors.WrapBoundary("mongodb", "find documents", err)
	}

	return DocumentResult{
		Found:     len(documents) > 0,
		Documents: documents,
		Count:     int64(len(documents)),
	}, nil
}

// Count counts matching documents with core-owned options.
func (m *MongoClient) Count(ctx context.Context, options CountOptions) (DocumentResult, error) {
	count, err := m.CountDocuments(ctx, options.Collection, options.Filter)
	if err != nil {
		return DocumentResult{}, coreerrors.WrapBoundary("mongodb", "count documents", err)
	}

	return DocumentResult{Found: count > 0, Count: count}, nil
}

// InsertDocument inserts one document with core-owned options.
func (m *MongoClient) InsertDocument(ctx context.Context, options InsertOptions) (WriteResult, error) {
	result, err := m.Collection(options.Collection).InsertOne(ctx, options.Document)
	if err != nil {
		return WriteResult{}, coreerrors.WrapBoundary("mongodb", "insert document", err)
	}

	return WriteResult{InsertedCount: 1, InsertedID: result.InsertedID}, nil
}

// UpdateDocuments updates one or many documents with core-owned options.
func (m *MongoClient) UpdateDocuments(ctx context.Context, options UpdateOptions) (WriteResult, error) {
	updateOptions := mongooptions.Update().SetUpsert(options.Upsert)

	if options.Many {
		result, err := m.Collection(options.Collection).UpdateMany(
			ctx,
			options.Filter,
			normalizeUpdate(options.Update),
			updateOptions,
		)
		if err != nil {
			return WriteResult{}, coreerrors.WrapBoundary("mongodb", "update documents", err)
		}

		return WriteResult{
			MatchedCount:  result.MatchedCount,
			ModifiedCount: result.ModifiedCount,
			UpsertedCount: result.UpsertedCount,
			UpsertedID:    result.UpsertedID,
		}, nil
	}

	result, err := m.Collection(options.Collection).UpdateOne(
		ctx,
		options.Filter,
		normalizeUpdate(options.Update),
		updateOptions,
	)
	if err != nil {
		return WriteResult{}, coreerrors.WrapBoundary("mongodb", "update document", err)
	}

	return WriteResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// DeleteDocuments deletes one or many documents with core-owned options.
func (m *MongoClient) DeleteDocuments(ctx context.Context, options DeleteOptions) (WriteResult, error) {
	if options.Many {
		result, err := m.Collection(options.Collection).DeleteMany(ctx, options.Filter)
		if err != nil {
			return WriteResult{}, coreerrors.WrapBoundary("mongodb", "delete documents", err)
		}

		return WriteResult{DeletedCount: result.DeletedCount}, nil
	}

	result, err := m.Collection(options.Collection).DeleteOne(ctx, options.Filter)
	if err != nil {
		return WriteResult{}, coreerrors.WrapBoundary("mongodb", "delete document", err)
	}

	return WriteResult{DeletedCount: result.DeletedCount}, nil
}

func (o FindOptions) findOptions() *mongooptions.FindOptions {
	options := mongooptions.Find()

	if o.Limit > 0 {
		options.SetLimit(o.Limit)
	}

	if o.Skip > 0 {
		options.SetSkip(o.Skip)
	}

	if o.Sort != nil {
		options.SetSort(o.Sort)
	}

	return options
}

func (o FindOptions) findOneOptions() *mongooptions.FindOneOptions {
	options := mongooptions.FindOne()

	if o.Skip > 0 {
		options.SetSkip(o.Skip)
	}

	if o.Sort != nil {
		options.SetSort(o.Sort)
	}

	return options
}

func normalizeUpdate(update any) any {
	if _, ok := update.(bson.D); ok {
		return update
	}

	if _, ok := update.(bson.M); ok {
		return update
	}

	return bson.D{{Key: "$set", Value: update}}
}
