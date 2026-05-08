package concurrent

import "testing"

func TestConcurrentContainers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "list stores values",
			run: func(t *testing.T) {
				t.Helper()

				list := NewList(1)
				list.Add(2)

				got := list.List()
				if len(got) != 2 {
					t.Fatalf("len = %d, want 2", len(got))
				}
			},
		},
		{
			name: "map stores values",
			run: func(t *testing.T) {
				t.Helper()

				items := NewMap(map[string]int{"a": 1})
				items.Set("b", 2)

				got, ok := items.Get("b")
				if !ok {
					t.Fatal("key b should exist")
				}

				if got != 2 {
					t.Fatalf("value = %d, want 2", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
