package metrics //nolint:revive // package name matches directory/domain usage

import (
	"errors"
	"testing"
)

func TestNew_Disabled(t *testing.T) {
	cfg := Config{}

	c, err := New(cfg, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if c != nil {
		t.Fatal("expected nil client when disabled")
	}
}

func TestNew_EnabledWithRegisteredProcessor(t *testing.T) {
	const kind = "test-enabled"

	Register(kind, func(_ Config, service string) (Processor, error) {
		return &spyProcessor{service: service}, nil
	})

	c, err := New(Config{Processors: []string{kind}}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if c == nil {
		t.Fatal("expected non-nil client when enabled")
	}

	if len(c.processors) != 1 {
		t.Fatalf("expected 1 processor, got %d", len(c.processors))
	}

	spy, ok := c.processors[0].(*spyProcessor)
	if !ok {
		t.Fatalf("unexpected processor type %T", c.processors[0])
	}

	if spy.service != "test-svc" {
		t.Fatalf("service = %q, want test-svc", spy.service)
	}
}

func TestRegistryNew(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*Registry)
		cfg     Config
		wantLen int
		wantErr bool
	}{
		{
			name: "explicit registry creates client",
			setup: func(registry *Registry) {
				registry.Register("explicit", func(_ Config, service string) (Processor, error) {
					return &spyProcessor{service: service}, nil
				})
			},
			cfg:     Config{Processors: []string{"explicit"}},
			wantLen: 1,
		},
		{
			name:    "missing processor returns error",
			setup:   func(*Registry) {},
			cfg:     Config{Processors: []string{"missing-explicit"}},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			registry := NewRegistry()
			test.setup(registry)

			client, err := registry.New(test.cfg, "test-svc")
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("registry new: %v", err)
			}

			if len(client.processors) != test.wantLen {
				t.Fatalf("processors = %d, want %d", len(client.processors), test.wantLen)
			}
		})
	}
}

func TestNew_UnregisteredProcessor(t *testing.T) {
	_, err := New(Config{Processors: []string{"missing-test-processor"}}, "test-svc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClose_NilReceiver(t *testing.T) {
	var c *Client

	if err := c.Close(); err != nil {
		t.Fatalf("Close() on nil client should return nil, got: %v", err)
	}
}

func TestCount_NilReceiver(t *testing.T) {
	var c *Client

	if err := c.Count("notifications_retention_deleted_total", 1, nil); err != nil {
		t.Fatalf("Count() on nil client should return nil, got: %v", err)
	}
}

func TestClientFansOut(t *testing.T) {
	first := &spyProcessor{}
	second := &spyProcessor{}
	c := &Client{processors: []Processor{first, second}, service: "test-svc"}

	if err := c.Count("count_total", 2, []string{"dry_run:false"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := c.Gauge("active_connections", 3, nil); err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := c.Distribution("wait_seconds", 1.5, nil); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}

	for _, processor := range []*spyProcessor{first, second} {
		processor.requireCall(t, "count:count_total")
		processor.requireCall(t, "gauge:active_connections")
		processor.requireCall(t, "distribution:wait_seconds")
	}
}

func TestClientReturnsProcessorErrors(t *testing.T) {
	c := &Client{processors: []Processor{&spyProcessor{err: errors.New("processor failed")}}}

	if err := c.Count("count_total", 1, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultClient(t *testing.T) {
	t.Cleanup(func() {
		SetDefault(nil)
	})

	c := &Client{processors: []Processor{&spyProcessor{}}, service: "test-svc"}
	SetDefault(c)

	if Default() != c {
		t.Fatal("expected default metrics client")
	}
}

func TestInstallDefault(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "restores previous default"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			previous := &Client{processors: []Processor{&spyProcessor{}}, service: "previous"}
			next := &Client{processors: []Processor{&spyProcessor{}}, service: "next"}

			SetDefault(previous)
			t.Cleanup(func() {
				SetDefault(nil)
			})

			handle := InstallDefault(next)
			if Default() != next {
				t.Fatal("default was not installed")
			}

			if err := handle.Close(); err != nil {
				t.Fatalf("close default handle: %v", err)
			}

			if Default() != previous {
				t.Fatal("previous default was not restored")
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	got := TagSet([]string{"status:200", "method:GET"})
	if got != "method:GET,status:200" {
		t.Fatalf("TagSet() = %q", got)
	}
}

type spyProcessor struct {
	service string
	err     error
	calls   []string
}

func (s *spyProcessor) Close() error {
	s.calls = append(s.calls, "close")

	return s.err
}

func (s *spyProcessor) Count(name string, _ int64, _ []string) error {
	s.calls = append(s.calls, "count:"+name)

	return s.err
}

func (s *spyProcessor) Gauge(name string, _ float64, _ []string) error {
	s.calls = append(s.calls, "gauge:"+name)

	return s.err
}

func (s *spyProcessor) Distribution(name string, _ float64, _ []string) error {
	s.calls = append(s.calls, "distribution:"+name)

	return s.err
}

func (s *spyProcessor) requireCall(t *testing.T, want string) {
	t.Helper()

	for _, call := range s.calls {
		if call == want {
			return
		}
	}

	t.Fatalf("missing call %q in %+v", want, s.calls)
}
