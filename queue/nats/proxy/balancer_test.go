package proxy

import (
	"log/slog"
	"testing"

	"github.com/InsideGallery/core/queue/nats/proxy/storage"
	"github.com/InsideGallery/core/testutils"
)

func TestBalancer(t *testing.T) {
	b := NewBalancer(storage.NewMemory())

	id := "29b710fcdbcb"
	instanceID1 := "abc1"
	instanceID2 := "abc2"
	instanceID3 := "abc3"

	slog.Default().Info("Add instance", "instance", instanceID1)
	testutils.Equal(t, b.AddInstance("test-subject", instanceID1), nil)

	slog.Default().Info("Execute balancer", "id", id)
	instance, err := b.Execute("test-subject", id)
	testutils.Equal(t, err, nil)
	slog.Default().Info("Choose instance", "instance", instance)

	slog.Default().Info("Add instance", "instance", instanceID2)
	testutils.Equal(t, b.AddInstance("test-subject", instanceID2), nil)

	slog.Default().Info("Execute balancer", "id", id)
	instance, err = b.Execute("test-subject", id)
	testutils.Equal(t, err, nil)
	slog.Default().Info("Choose instance", "instance", instance)

	slog.Default().Info("Add instance", "instance", instanceID3)
	testutils.Equal(t, b.AddInstance("test-subject", instanceID3), nil)

	slog.Default().Info("Execute balancer", "id", id)
	instance, err = b.Execute("test-subject", id)
	testutils.Equal(t, err, nil)
	slog.Default().Info("Choose instance", "instance", instance)

	slog.Default().Info("Destroy instance", "instance", instanceID1)
	testutils.Equal(t, b.DestroyInstance(instanceID1), nil)

	slog.Default().Info("Execute balancer", "id", id)
	instance, err = b.Execute("test-subject", id)
	testutils.Equal(t, err, nil)
	slog.Default().Info("Choose instance", "instance", instance)
}
