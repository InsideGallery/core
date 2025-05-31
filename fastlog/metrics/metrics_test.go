//go:build local_test
// +build local_test

package metrics

import (
	"context"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestMetrics(t *testing.T) {
	m, err := Default(context.Background())
	defer m.Shutdown()

	testutils.Equal(t, err, nil)
	h, err := m.Histogram("testchart")
	testutils.Equal(t, err, nil)
	err = h.Execute(context.Background(), 1, "calls")
	err = h.Execute(context.Background(), 5, "calls1")
	err = h.Execute(context.Background(), 10, "calls2")
	err = h.Execute(context.Background(), 1, "calls3")
	err = h.Execute(context.Background(), 1, "calls3")
	testutils.Equal(t, err, nil)
}
