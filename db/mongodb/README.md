# db/mongodb

Import path: `github.com/InsideGallery/core/db/mongodb`

Package `mongodb` provides MongoDB client construction, client ownership helpers, document operations,
and filter/document builders. New code should use `DocumentStore` and core-owned option/result types
when application code should not expose MongoDB SDK result types.

## Main APIs

- `ConnectionConfig` configures hosts, scheme, database, credentials, read preference, retry writes, and
  URI arguments.
- `GetConnectionConfigFromEnv()` reads `MONGO_*` environment variables.
- `NewMongoClient(config)` creates a `MongoClient`.
- `ClientStore` owns a `MongoClient` for explicit application composition.
- `DocumentStore` is the core-owned interface for find, count, insert, update, and delete operations.
- `FindOptions`, `CountOptions`, `InsertOptions`, `UpdateOptions`, and `DeleteOptions` describe document
  operations.
- `DocumentResult` and `WriteResult` return core-owned read and write results.
- `Field`, `Filter`, `Document`, `SortField`, `NewFilter`, `NewDocument`, `FilterFromPairs`,
  `DocumentFromPairs`, and `NewSort` build MongoDB-compatible inputs.
- `Client`, `Set`, `Get`, `Default`, and `GetBsonD` are legacy compatibility APIs.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/mongodb"
)

type Profile struct {
	Name string `bson:"name"`
}

func findProfile(ctx context.Context, config *mongodb.ConnectionConfig) (profile Profile, err error) {
	client, err := mongodb.NewMongoClient(config)
	if err != nil {
		return Profile{}, err
	}
	defer func() {
		if closeErr := client.Disconnect(ctx); err == nil {
			err = closeErr
		}
	}()

	_, err = client.FindOneDocument(ctx, mongodb.FindOptions{
		Collection: "profiles",
		Filter:     mongodb.NewFilter(mongodb.Field{Name: "name", Value: "Ada"}),
		Target:     &profile,
	})

	return profile, err
}
```

## Configuration And Operations

Environment variables include `MONGO_HOSTS`, `MONGO_USER`, `MONGO_PASS`, `MONGO_SCHEME`,
`MONGO_DATABASE`, `MONGO_ARGS`, `MONGO_MODE`, `MONGO_RETRYWRITES`, `MONGO_AUTH_MECHANISM`, and
`MONGO_AUTH_SOURCE`. `ConnectionConfig.GetDSN()` builds `scheme://host1,host2/?args`; authentication is
attached separately when user and password are both set. `FindOneDocument` and `FindDocuments` require a
non-nil target. Updates are wrapped in `$set` unless the update is already a `bson.D` or `bson.M`.
