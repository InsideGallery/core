package mongodb

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClient client for mongo db
type MongoClient struct {
	*mongo.Client
	database string
}

// NewMongoClient return client from config
func NewMongoClient(config *ConnectionConfig) (*MongoClient, error) {
	ctx := context.Background()

	cs := options.Client().ApplyURI(config.GetDSN()).SetRetryWrites(config.RetryWrites)

	if config.User != "" && config.Pass != "" {
		credential := options.Credential{
			AuthMechanism: config.AuthMechanism,
			AuthSource:    config.AuthSource,
			Username:      config.User,
			Password:      config.Pass,
		}
		cs = cs.SetAuth(credential)
	}

	mode, err := readpref.ModeFromString(config.Mode)
	if err != nil {
		mode = readpref.SecondaryPreferredMode
	}

	pref, err := readpref.New(mode)
	if err != nil {
		cs = cs.SetReadPreference(pref)
	}

	client, err := mongo.Connect(ctx, cs)
	if err != nil {
		return nil, err
	}

	return &MongoClient{
		database: config.Database,
		Client:   client,
	}, nil
}

func (m *MongoClient) WithDB(name string) Client {
	return &MongoClient{
		database: name,
		Client:   m.Client,
	}
}

func (m *MongoClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	return m.Client.Database(name, opts...)
}

func (m *MongoClient) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return m.Client.Database(m.database).Collection(name, opts...)
}

// FindOne find and decode one element
func (m *MongoClient) FindOne(
	ctx context.Context,
	collection string,
	value interface{},
	filter interface{},
	opts ...*options.FindOneOptions,
) error {
	result := m.Collection(collection).FindOne(ctx, filter, opts...)
	return result.Decode(value)
}

func (m *MongoClient) retrieve(ctx context.Context, cur *mongo.Cursor, value interface{}) ([]interface{}, error) {
	var err error

	defer func() {
		if closeErr := cur.Close(ctx); closeErr != nil {
			err = closeErr
		}
	}()

	var result []interface{}

	for cur.Next(ctx) {
		valueLocal := Clone(value)

		err = cur.Decode(valueLocal)
		if err != nil {
			return nil, err
		}

		v := reflect.Indirect(reflect.ValueOf(valueLocal))
		item := v.Interface()

		result = append(result, item)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}

	return result, err
}

func Clone(oldObj interface{}) interface{} {
	newObj := reflect.New(reflect.TypeOf(oldObj).Elem())
	oldVal := reflect.ValueOf(oldObj).Elem()
	newVal := newObj.Elem()

	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return newObj.Interface()
}

// Find all and return
func (m *MongoClient) Find(
	ctx context.Context,
	collection string,
	value interface{},
	filter interface{},
	opts ...*options.FindOptions,
) ([]interface{}, error) {
	cur, err := m.Collection(collection).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return m.retrieve(ctx, cur, value)
}

// FindOneByID return value by _id
func (m *MongoClient) FindOneByID(
	ctx context.Context,
	collection string,
	value interface{},
	id interface{},
	opts ...*options.FindOneOptions,
) error {
	filter := bson.D{{Key: "_id", Value: id}}
	result := m.Collection(collection).FindOne(ctx, filter, opts...)

	return result.Decode(value)
}

// CountDocuments count documents
func (m *MongoClient) CountDocuments(
	ctx context.Context,
	collection string,
	filter interface{},
	opts ...*options.CountOptions,
) (int64, error) {
	count, err := m.Collection(collection).CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Aggregate aggregate rows
func (m *MongoClient) Aggregate(
	ctx context.Context,
	collection string,
	value interface{},
	pipeline interface{},
	opts ...*options.AggregateOptions,
) ([]interface{}, error) {
	cur, err := m.Collection(collection).Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}

	return m.retrieve(ctx, cur, value)
}

// InsertOne insert single object
func (m *MongoClient) InsertOne(
	ctx context.Context,
	collection string,
	value interface{},
	opts ...*options.InsertOneOptions,
) error {
	_, err := m.Collection(collection).InsertOne(ctx, value, opts...)
	return err
}

// UpsertOne insert or update single object
func (m *MongoClient) UpsertOne(
	ctx context.Context,
	collection string,
	update interface{},
	filter interface{},
	opts ...*options.UpdateOptions,
) error {
	o := options.Update()
	o.SetUpsert(true)
	opts = append(opts, o)
	_, err := m.Collection(collection).UpdateOne(ctx, filter, update, opts...)

	return err
}

// UpsertMany insert or update objects
func (m *MongoClient) UpsertMany(
	ctx context.Context,
	collection string,
	keys []interface{},
	documents []interface{},
) error {
	models := make([]mongo.WriteModel, 0, len(keys))

	for i, key := range keys {
		models = append(
			models,
			mongo.
				NewUpdateOneModel().
				SetFilter(bson.D{{Key: "_id", Value: key}}).
				SetUpdate(bson.D{{Key: "$set", Value: documents[i]}}).
				SetUpsert(true),
		)
	}

	opts := options.BulkWrite().SetOrdered(false)

	_, err := m.Collection(collection).BulkWrite(ctx, models, opts)
	if err != nil {
		return err
	}

	return nil
}

// UpsertManyByFilter insert or update objects by filter
func (m *MongoClient) UpsertManyByFilter(
	ctx context.Context,
	collection string,
	filterBy string,
	keys []interface{},
	documents []interface{},
) error {
	models := make([]mongo.WriteModel, 0, len(keys))

	for i, key := range keys {
		models = append(
			models,
			mongo.
				NewUpdateOneModel().
				SetFilter(bson.D{{Key: filterBy, Value: key}}).
				SetUpdate(bson.D{{Key: "$set", Value: documents[i]}}).
				SetUpsert(true),
		)
	}

	opts := options.BulkWrite().SetOrdered(false)

	_, err := m.Collection(collection).BulkWrite(ctx, models, opts)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) BatchUpdateByID(ctx context.Context, collection string, data map[interface{}]interface{}) error {
	models := make([]mongo.WriteModel, 0, len(data))

	for id, v := range data {
		models = append(
			models,
			mongo.
				NewUpdateOneModel().
				SetFilter(bson.D{{Key: "_id", Value: id}}).
				SetUpdate(bson.D{{Key: "$set", Value: v}}),
		)
	}

	opts := options.BulkWrite().SetOrdered(true)

	_, err := m.Collection(collection).BulkWrite(ctx, models, opts)
	if err != nil {
		return err
	}

	return nil
}

// UpdateByObject update to an object by filter
func (m *MongoClient) UpdateByObject(
	ctx context.Context,
	collection string,
	value interface{},
	filter interface{},
	opts ...*options.UpdateOptions,
) error {
	o := options.Update()
	o.SetUpsert(true)
	opts = append(opts, o)
	update := bson.D{{Key: "$set", Value: value}}
	_, err := m.Collection(collection).UpdateOne(ctx, filter, update, opts...)

	return err
}

// DeleteOne delete one element
func (m *MongoClient) DeleteOne(
	ctx context.Context,
	collection string,
	filter interface{},
	opts ...*options.DeleteOptions,
) error {
	_, err := m.Collection(collection).DeleteOne(ctx, filter, opts...)
	return err
}

// DeleteMany delete all elements by condition
func (m *MongoClient) DeleteMany(
	ctx context.Context,
	collection string,
	filter interface{},
	opts ...*options.DeleteOptions,
) error {
	_, err := m.Collection(collection).DeleteMany(ctx, filter, opts...)
	return err
}

// Drop collection
func (m *MongoClient) Drop(ctx context.Context, collection string) error {
	return m.Collection(collection).Drop(ctx)
}

func (m *MongoClient) Connection() *mongo.Client {
	return m.Client
}

// InsertMany insert many documents
func (m *MongoClient) InsertMany(
	ctx context.Context,
	collection string,
	documents []interface{},
	opts ...*options.InsertManyOptions,
) error {
	o := options.InsertMany()
	o.SetOrdered(false)
	opts = append(opts, o)
	_, err := m.Collection(collection).InsertMany(ctx, documents, opts...)

	return err
}

func (m *MongoClient) DeleteCollection(ctx context.Context, name string) error {
	return m.Database(m.database).Collection(name).Drop(ctx)
}
