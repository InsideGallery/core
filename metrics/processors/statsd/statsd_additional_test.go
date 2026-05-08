package statsd

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/InsideGallery/core/metrics"
)

func TestProcessorWritesStatsDPackets(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer listener.Close()

	t.Setenv("METRICS_STATSD_ADDR", listener.LocalAddr().String())
	t.Setenv("METRICS_STATSD_NAMESPACE", "svc")

	rawProcessor, err := New(metrics.Config{}, "unit")
	if err != nil {
		t.Fatalf("new statsd processor: %v", err)
	}
	defer func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("close processor: %v", err)
		}
	}()

	cases := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "count",
			run: func() error {
				return rawProcessor.Count("requests.total", 3, nil)
			},
			want: "svc.requests.total:3|c",
		},
		{
			name: "gauge sanitizes name",
			run: func() error {
				return rawProcessor.Gauge("queue depth", 12.5, nil)
			},
			want: "svc.queue_depth:12.5|g",
		},
		{
			name: "distribution defaults empty name",
			run: func() error {
				return rawProcessor.Distribution("", 7, nil)
			},
			want: "svc.metric:7|ms",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err != nil {
				t.Fatalf("write metric: %v", err)
			}

			got := readStatsDPacket(t, listener)
			if got != test.want {
				t.Fatalf("packet = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewRequiresAddress(t *testing.T) {
	t.Setenv("METRICS_STATSD_ADDR", "")

	processor, err := New(metrics.Config{}, "unit")
	if err == nil {
		t.Fatal("expected error")
	}

	if processor != nil {
		t.Fatalf("processor = %#v, want nil", processor)
	}
}

func TestProcessorWriteError(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer listener.Close()

	t.Setenv("METRICS_STATSD_ADDR", listener.LocalAddr().String())

	rawProcessor, err := New(metrics.Config{}, "unit")
	if err != nil {
		t.Fatalf("new statsd processor: %v", err)
	}

	if err := rawProcessor.Close(); err != nil {
		t.Fatalf("close processor: %v", err)
	}

	if err := rawProcessor.Count("closed", 1, nil); err == nil {
		t.Fatal("expected write error")
	}
}

func TestConfigNamespaceDefault(t *testing.T) {
	cases := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name: "empty namespace uses default",
			want: "ptolemy.",
		},
		{
			name:      "trims trailing dot",
			namespace: "custom.",
			want:      "custom.",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("METRICS_STATSD_NAMESPACE", test.namespace)

			cfg, err := getConfigFromEnv()
			if err != nil {
				t.Fatalf("get config: %v", err)
			}

			if got := cfg.namespacePrefix(); got != test.want {
				t.Fatalf("namespace prefix = %q, want %q", got, test.want)
			}
		})
	}
}

func readStatsDPacket(t *testing.T, listener net.PacketConn) string {
	t.Helper()

	if err := listener.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	buffer := make([]byte, 256)
	n, _, err := listener.ReadFrom(buffer)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			t.Fatalf("listener closed: %v", err)
		}

		t.Fatalf("read packet: %v", err)
	}

	return string(buffer[:n])
}
