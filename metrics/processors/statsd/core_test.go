package statsd

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/InsideGallery/core/metrics"
)

func TestNewRequiresAddress(t *testing.T) {
	t.Setenv("METRICS_STATSD_ADDR", "")

	if _, err := New(metrics.Config{}, "test-svc"); err == nil {
		t.Fatal("expected missing address error")
	}
}

func TestProcessorWritesStatsDPackets(t *testing.T) {
	listenConfig := net.ListenConfig{}

	conn, err := listenConfig.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error: %v", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("Close(listener) error: %v", err)
		}
	}()

	t.Setenv("METRICS_STATSD_ADDR", conn.LocalAddr().String())
	t.Setenv("METRICS_STATSD_NAMESPACE", "custom")

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	defer func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close(processor) error: %v", err)
		}
	}()

	if err := rawProcessor.Count("requests total", 2, nil); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := rawProcessor.Gauge("active connections", 3.5, nil); err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := rawProcessor.Distribution("wait seconds", 1.25, nil); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	packets := readPackets(t, conn, 3)
	for _, want := range []string{
		"custom.requests_total:2|c",
		"custom.active_connections:3.5|g",
		"custom.wait_seconds:1.25|ms",
	} {
		if !containsPacket(packets, want) {
			t.Fatalf("packets = %v, want %q", packets, want)
		}
	}
}

func TestWriteReturnsErrorAfterClose(t *testing.T) {
	listenConfig := net.ListenConfig{}

	conn, err := listenConfig.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error: %v", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("Close(listener) error: %v", err)
		}
	}()

	t.Setenv("METRICS_STATSD_ADDR", conn.LocalAddr().String())

	rawProcessor, err := New(metrics.Config{}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if err := rawProcessor.Close(); err != nil {
		t.Fatalf("Close(processor) error: %v", err)
	}

	err = rawProcessor.Count("requests", 1, nil)
	if err == nil {
		t.Fatal("expected write error")
	}

	if !strings.Contains(err.Error(), "write statsd packet") {
		t.Fatalf("Count() error = %v, want write context", err)
	}
}

func readPackets(t *testing.T, conn net.PacketConn, count int) []string {
	t.Helper()

	packets := make([]string, 0, count)
	buf := make([]byte, 256)

	for len(packets) < count {
		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatalf("SetReadDeadline() error: %v", err)
		}

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			t.Fatalf("ReadFrom() error after packets %v: %v", packets, err)
		}

		packets = append(packets, string(buf[:n]))
	}

	return packets
}

func containsPacket(packets []string, want string) bool {
	for _, packet := range packets {
		if packet == want {
			return true
		}
	}

	return false
}
