package registry

import "testing"

func TestGroup(t *testing.T) {
	r := NewRegistry[string, uint64, any]()

	data := []*MockEntity{
		{
			id: r.NextID(),
		},
		{
			id: r.NextID(),
		},
		{
			id: r.NextID(),
		},
	}

	for _, e := range data {
		err := r.GetGroup(KeyTemporary).Add(r.NextID(), e)
		if err != nil {
			t.Fatal(err)
		}
	}

	result := r.GetValues(KeyTemporary)

	if len(data) != len(result) {
		t.Fatalf("Unexpected objects count: %d != %d", len(data), len(result))
	}

	for _, item := range result {
		var exists bool

		e := item.(*MockEntity)
		for _, i := range data {
			if i == e {
				exists = true
			}
		}

		if !exists {
			t.Fatalf("Not found expected object: %+v", e)
		}
	}

	r.DeleteGroup(KeyTemporary)

	result = r.GetValues(KeyTemporary)
	if len(result) != 0 {
		t.Fatalf("Unexpected objects count: %d", len(result))
	}

	for _, e := range data {
		err := r.GetGroup(KeyTemporary).Add(r.NextID(), e)
		if err != nil {
			t.Fatal(err)
		}
	}

	r.TruncateGroup(KeyTemporary)

	result = r.GetValues(KeyTemporary)
	if len(result) != 0 {
		t.Fatalf("Unexpected objects count: %d", len(result))
	}

	res := r.SearchInGroup(KeyTemporary, func(_ interface{}, id interface{}, _ interface{}) bool {
		return id.(uint32) == 1
	})
	for e := range res {
		var exists bool

		entity := e.(*MockEntity)
		for _, i := range data {
			if i == entity {
				exists = true
			}
		}

		if !exists {
			t.Fatalf("Not found expected object: %+v", e)
		}
	}
}
