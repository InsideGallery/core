package worker

import (
	"testing"
	"time"
)

func TestMergeChannels(t *testing.T) {
	cases := []struct {
		name   string
		inputs []<-chan interface{}
		want   map[interface{}]bool
	}{
		{
			name: "no inputs closes output",
			want: map[interface{}]bool{},
		},
		{
			name: "merges closed inputs",
			inputs: []<-chan interface{}{
				closedInterfaceChannel("a", "b"),
				closedInterfaceChannel("c"),
			},
			want: map[interface{}]bool{
				"a": true,
				"b": true,
				"c": true,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := map[interface{}]bool{}
			for value := range MergeChannels(test.inputs...) {
				got[value] = true
			}

			if len(got) != len(test.want) {
				t.Fatalf("merged item count = %d, want %d", len(got), len(test.want))
			}

			for value := range test.want {
				if !got[value] {
					t.Fatalf("missing merged value %v", value)
				}
			}
		})
	}
}

func TestGetMessageOrTimeout(t *testing.T) {
	cases := []struct {
		name    string
		message []byte
		def     []byte
		want    []byte
	}{
		{
			name:    "message wins",
			message: []byte("payload"),
			def:     []byte("default"),
			want:    []byte("payload"),
		},
		{
			name: "timeout wins",
			def:  []byte("default"),
			want: []byte("default"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			messages := make(chan []byte, 1)
			if test.message != nil {
				messages <- test.message
			}

			got := GetMessageOrTimeout(time.Millisecond, messages, test.def)
			if string(got) != string(test.want) {
				t.Fatalf("GetMessageOrTimeout() = %q, want %q", got, test.want)
			}
		})
	}
}

func closedInterfaceChannel(values ...interface{}) <-chan interface{} {
	ch := make(chan interface{}, len(values))
	for _, value := range values {
		ch <- value
	}

	close(ch)

	return ch
}
