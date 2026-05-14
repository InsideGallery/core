// Package elasticsearch provides Elasticsearch search client helpers.
//
// New code should use the core-owned search boundary:
//
//	import "github.com/InsideGallery/core/db/elasticsearch"
//
//	client, err := elasticsearch.NewSearchClient(elasticsearch.Options{
//		Addresses: []strings{"http://localhost:9200"},
//	})
//
// Use Searcher, SearchOptions, and SearchResult for consumer-facing code that
// should not expose Elasticsearch SDK request or response types.
//
// Compatibility: Client and NewClient remain available for existing SDK-shaped
// callers. Prefer NewSearchClient for new integrations.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	coreerrors "github.com/InsideGallery/core/errors"
)

// Options is the core-owned input for creating an Elasticsearch search client.
type Options struct {
	Addresses []string
	Username  string
	Password  string
	CloudID   string
	APIKey    string
}

// SearchOptions is the core-owned input for an Elasticsearch search.
type SearchOptions struct {
	Indexes []string
	Query   map[string]any
}

// SearchResult is the core-owned search result.
type SearchResult struct {
	Body map[string]any
}

// Searcher is the core-owned Elasticsearch contract for new consumers.
type Searcher interface {
	Search(ctx context.Context, options SearchOptions) (SearchResult, error)
}

// SearchClient wraps the Elasticsearch SDK behind core-owned inputs and results.
type SearchClient struct {
	search esapi.Search
}

// NewSearchClient creates an Elasticsearch search client from core-owned options.
func NewSearchClient(options Options) (*SearchClient, error) {
	client, err := newElasticClient(options)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client: %w", err)
	}

	return &SearchClient{search: client.Search}, nil
}

// Search searches indexes through core-owned inputs and results.
func (c *SearchClient) Search(ctx context.Context, options SearchOptions) (SearchResult, error) {
	body, err := searchByIndex(ctx, c.search, options.Indexes, options.Query)
	if err != nil {
		return SearchResult{}, err
	}

	return SearchResult{Body: body}, nil
}

func newElasticClient(options Options) (*elasticsearch.Client, error) {
	if len(options.Addresses) == 0 &&
		options.Username == "" &&
		options.Password == "" &&
		options.CloudID == "" &&
		options.APIKey == "" {
		return elasticsearch.NewDefaultClient()
	}

	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses: options.Addresses,
		Username:  options.Username,
		Password:  options.Password,
		CloudID:   options.CloudID,
		APIKey:    options.APIKey,
	})
}

func searchByIndex(
	ctx context.Context,
	search esapi.Search,
	indexes []string,
	query map[string]any,
) (result map[string]any, err error) {
	result = map[string]any{}

	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return result, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := search(
		search.WithContext(ctx),
		search.WithIndex(indexes...),
		search.WithBody(&buf),
		search.WithTrackTotalHits(true),
		search.WithPretty(),
	)
	if err != nil {
		return result, coreerrors.WrapBoundary("elasticsearch", "search", err)
	}

	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("error closing response: %w", closeErr)
		}
	}()

	if res.IsError() {
		return result, searchResponseError(res)
	}

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error parsing the response body: %w", err)
	}

	return result, nil
}

func searchResponseError(res *esapi.Response) error {
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return fmt.Errorf("error parsing the response body: %w", err)
	}

	errorInfo, ok := body["error"].(map[string]any)
	if !ok {
		return fmt.Errorf("[%s]: %w", res.Status(), ErrWrongResponse)
	}

	errorType, _ := errorInfo["type"].(string)
	reason, _ := errorInfo["reason"].(string)

	return fmt.Errorf("[%s] %s: %s: %w", res.Status(), errorType, reason, ErrWrongResponse)
}
