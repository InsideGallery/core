# db/elasticsearch

Import path: `github.com/InsideGallery/core/db/elasticsearch`

Package `elasticsearch` wraps the Elasticsearch Go client with core-owned search inputs and results.
New application code should depend on `Searcher` instead of exposing Elasticsearch SDK request and
response types at application boundaries.

## Main APIs

- `Options` configures addresses, username/password, Cloud ID, and API key.
- `NewSearchClient(options)` creates a `SearchClient`.
- `Searcher` is the core-owned interface implemented by `SearchClient`.
- `SearchOptions` supplies index names and a JSON-serializable query map.
- `SearchResult` returns the decoded response body as `map[string]any`.
- `ErrWrongResponse` reports Elasticsearch error responses that cannot be treated as successful search
  results.
- `Client`, `NewClient`, `GetMatchQuery`, and `SearchByIndex` are legacy SDK-shaped APIs. New code should
  prefer `SearchClient`.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/db/elasticsearch"
)

func search(ctx context.Context) (elasticsearch.SearchResult, error) {
	client, err := elasticsearch.NewSearchClient(elasticsearch.Options{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		return elasticsearch.SearchResult{}, err
	}

	return client.Search(ctx, elasticsearch.SearchOptions{
		Indexes: []string{"products"},
		Query: map[string]any{
			"query": map[string]any{"match_all": map[string]any{}},
		},
	})
}
```

## Configuration And Operations

An empty `Options` value uses the Elasticsearch SDK default client. `Search` encodes the query as JSON,
enables total-hit tracking, decodes the response body into a map, and closes the response body. Transport
errors and Elasticsearch error responses are returned as errors.
