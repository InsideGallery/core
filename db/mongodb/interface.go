//go:generate mockgen -source=interface.go -destination=mocks/client.go
package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client interface {
	FindOne(
		ctx context.Context,
		collection string, value interface{}, filter interface{}, opts ...*options.FindOneOptions) error
	Find(
		ctx context.Context,
		collection string,
		value interface{},
		filter interface{},
		opts ...*options.FindOptions,
	) ([]interface{}, error)
	FindOneByID(
		ctx context.Context,
		collection string,
		value interface{},
		id interface{},
		opts ...*options.FindOneOptions,
	) error
	CountDocuments(
		ctx context.Context,
		collection string,
		filter interface{},
		opts ...*options.CountOptions,
	) (int64, error)
	Aggregate(
		ctx context.Context,
		collection string,
		value interface{},
		pipeline interface{},
		opts ...*options.AggregateOptions,
	) ([]interface{}, error)
	InsertOne(
		ctx context.Context,
		collection string,
		value interface{},
		opts ...*options.InsertOneOptions,
	) error
	UpsertOne(
		ctx context.Context,
		collection string,
		update interface{},
		filter interface{},
		opts ...*options.UpdateOptions,
	) error
	UpsertMany(
		ctx context.Context,
		collection string,
		keys []interface{},
		documents []interface{},
	) error
	UpsertManyByFilter(
		ctx context.Context,
		collection string,
		filterBy string,
		keys []interface{},
		documents []interface{},
	) error
	UpdateByObject(
		ctx context.Context,
		collection string,
		value interface{},
		filter interface{},
		opts ...*options.UpdateOptions,
	) error
	DeleteOne(ctx context.Context, collection string, filter interface{}, opts ...*options.DeleteOptions) error
	DeleteMany(ctx context.Context, collection string, filter interface{}, opts ...*options.DeleteOptions) error
	Drop(ctx context.Context, collection string) error
	InsertMany(ctx context.Context, collection string, documents []interface{}, opts ...*options.InsertManyOptions) error
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
	Database(name string, opts ...*options.DatabaseOptions) *mongo.Database
	WithDB(name string) Client
	Connection() *mongo.Client
	BatchUpdateByID(ctx context.Context, collection string, data map[interface{}]interface{}) error
	DeleteCollection(ctx context.Context, name string) error
}
