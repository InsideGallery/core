package throughput

import (
	"context"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestThroughput(t *testing.T) {
	storage := NewMemoryStorage()

	th := New(context.Background(), storage)
	go th.Loop()

	var statuses []bool

	for i := 0; i < int(Tier0RPM+1); i++ {
		res := th.Validate("test")
		statuses = append(statuses, res)
	}

	testutils.Equal(t, len(statuses) > 1, true)
	testutils.Equal(t, statuses[0], true)
	testutils.Equal(t, statuses[len(statuses)-1], false)
}

var statusesL []bool

func BenchmarkThroughput(b *testing.B) {
	storage := NewMemoryStorage()

	th := New(context.Background(), storage)
	go th.Loop()

	for i := 0; i < b.N; i++ {
		res := th.Validate("test")
		statusesL = append(statusesL, res)
	}
}
