package order

import "testing"

func TestSort(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []interface{}
		want []int
	}{
		{
			name: "sort integers",
			data: []interface{}{3, 1, 2},
			want: []int{1, 2, 3},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			Sort(test.data, func(a, b interface{}) int {
				left := a.(int)
				right := b.(int)

				return left - right
			})

			for i, want := range test.want {
				got := test.data[i].(int)
				if got != want {
					t.Fatalf("data[%d] = %d, want %d", i, got, want)
				}
			}
		})
	}
}
