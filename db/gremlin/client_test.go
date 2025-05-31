package gremlin

import (
	"fmt"
	"testing"

	"github.com/InsideGallery/core/testutils"
	"github.com/InsideGallery/core/utils"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

func TestGremlin(t *testing.T) {
	t.Skip() // only for understand how it works

	cfg, err := GetConnectionConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	// cfg.URL = "ws://10.173.84.182:8182/gremlin" // for debug purpose
	// cfg.URL = "wss://127.0.0.1:8182/gremlin" // for debug purpose
	c, err := New(cfg, func(settings *gremlingo.DriverRemoteConnectionSettings) {
		settings.TlsConfig.InsecureSkipVerify = true
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	id1, err := utils.GetShortID()
	testutils.Equal(t, err, nil)
	id2, err := utils.GetShortID()
	testutils.Equal(t, err, nil)

	fmt.Println(string(id1), string(id2))
	id1 = []byte("34f30e6d69bb")
	id2 = []byte("38220e6d9f4e")

	cache := NewCache()
	defer cache.Truncate()

	ops := []Operation{
		NewUpsertVertexOp("airports", string(id1)),
		NewUpsertVertexOp("airports", string(id2), "age", 2),
		NewUpsertEdgeOp("route", string(append(id1, id2...)),
			NewLabelVertexGetter("airports", string(id1)),
			NewLabelVertexGetter("airports", string(id2)),
		),
		NewCallbackOp(func(cache *Cache, source *gremlingo.GraphTraversalSource) ([]*gremlingo.Result, error) {
			fmt.Println(cache.Registry.GetGroups())

			src := source.GetGraphTraversal()

			src = MergeV(src, "test_merge2", "id1", map[interface{}]interface{}{
				"field": "max",
			})
			res, err := MergeV(src, "test_merge2", "id2", map[interface{}]interface{}{
				"field":  "max",
				"field2": "max2",
			}).Next()
			if err != nil {
				return nil, err
			}
			fmt.Println(res)

			getter1 := NewLabelVertexGetter("test_merge2", "id1")
			getter2 := NewLabelVertexGetter("test_merge2", "id2")

			id1, v1, err := getter1.Get(cache, source)
			if err != nil {
				return nil, err
			}
			id2, v2, err := getter2.Get(cache, source)
			if err != nil {
				return nil, err
			}
			fmt.Println(id1, id2)

			src = source.GetGraphTraversal()
			res, err = MergeE(src, "test_edge2", "123", v1, v2, map[interface{}]interface{}{
				"eopt": 123,
			}).Next()
			if err != nil {
				return nil, err
			}
			fmt.Println(res)
			return nil, nil
		}),
		// NewDropVertexOp(NewLabelVertexGetter("airports", "TST")),
		// NewDropVertexOp(NewLabelVertexGetter("airports", "TST2")),
	}

	testutils.Equal(t, c.Execute(cache, ops...), nil)

	ops = []Operation{
		NewUpsertVertexOp("airports", string(id1), "age", 4),
		NewUpsertVertexOp("airports", string(id2), "age", 3),
		// NewDropVertexOp(NewLabelVertexGetter("airports", "TST")),
		// NewDropVertexOp(NewLabelVertexGetter("airports", "TST2")),
	}

	testutils.Equal(t, c.Execute(cache, ops...), nil)
}
