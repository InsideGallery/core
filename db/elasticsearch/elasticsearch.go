package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

func prepareKeyValues(keyValues ...string) (map[string][]string, error) {
	l := len(keyValues)
	if l == 0 {
		return nil, nil
	}

	if l%2 != 0 {
		return nil, ErrWrongCountOfArguments
	}

	attributes := map[string][]string{}

	for i := 0; i < len(keyValues)-1; i += 2 {
		key, value := keyValues[i], keyValues[i+1]
		attributes[key] = append(attributes[key], value)
	}

	return attributes, nil
}

// Client is the legacy Elasticsearch SDK-shaped client.
//
// Deprecated: use SearchClient and core-owned option/result types for new code.
type Client struct {
	*elasticsearch.Client
}

// NewClient creates the legacy Elasticsearch SDK-shaped client.
//
// Deprecated: use NewSearchClient for new code.
func NewClient() (*Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("get response: %w", err)
	}

	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("close response: %w", err)
	}

	// Check response status
	if res.IsError() {
		return nil, fmt.Errorf("response info: %w", ErrWrongResponse)
	}

	return &Client{
		Client: es,
	}, nil
}

func (c *Client) GetMatchQuery(keyValues ...string) (map[string]interface{}, error) {
	v, err := prepareKeyValues(keyValues...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"match": v,
		},
	}, nil
}

func (c *Client) SearchByIndex(
	ctx context.Context,
	indexes []string,
	query map[string]interface{},
) (map[string]interface{}, error) {
	return searchByIndex(ctx, c.Search, indexes, query)
}
