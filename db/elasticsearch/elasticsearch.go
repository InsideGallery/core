package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
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

type Client struct {
	*elasticsearch.Client
}

func NewClient() (*Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("error creating the client: %w", err)
	}

	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting responset: %w", err)
	}

	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error close responset: %w", err)
	}

	// Check response status
	if res.IsError() {
		return nil, fmt.Errorf("error response info: %w", ErrWrongResponse)
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
	r := map[string]interface{}{}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return r, fmt.Errorf("error encoding query: %w", err)
	}

	// Perform the search request.
	res, err := c.Client.Search(
		c.Client.Search.WithContext(ctx),
		c.Client.Search.WithIndex(indexes...),
		c.Client.Search.WithBody(&buf),
		c.Client.Search.WithTrackTotalHits(true),
		c.Client.Search.WithPretty(),
	)
	if err != nil {
		return r, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return r, fmt.Errorf("error parsing the response body: %w", err)
		}

		return r, fmt.Errorf("[%s] %s: %s: %w",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
			ErrWrongResponse,
		)
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return r, fmt.Errorf("error parsing the response body: %w", err)
	}

	return r, nil
}
