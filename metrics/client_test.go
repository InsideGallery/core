package metrics //nolint:revive // package name matches directory/domain usage

import (
	"errors"
	"sort"
	"strings"
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

func TestNew_UnregisteredProcessor(t *testing.T) {
	_, err := New(Config{Processors: []string{"missing-test-processor"}}, "test-svc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNew_ReturnsFactoryError(t *testing.T) {
	kind := "test-factory-error"
	wantErr := errors.New("factory failed")

	Register(kind, func(_ Config, _ string) (Processor, error) {
		return nil, wantErr
	})

	_, err := New(Config{Processors: []string{kind}}, "test-svc")
	if !errors.Is(err, wantErr) {
		t.Fatalf("New() error = %v, want %v", err, wantErr)
	}

	if !strings.Contains(err.Error(), kind) {
		t.Fatalf("New() error = %q, want processor kind", err.Error())
	}
}

func TestNew_SkipsNilProcessor(t *testing.T) {
	kind := "test-nil-processor"

	Register(kind, func(_ Config, _ string) (Processor, error) {
		return nil, nil
	})

	c, err := New(Config{Processors: []string{kind}}, "test-svc")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if c != nil {
		t.Fatal("expected nil client when all processors are no-op")
	}
}

func TestRegisteredProcessorsReturnsSortedNormalizedNames(t *testing.T) {
	Register(" Zz-Unit ", func(_ Config, _ string) (Processor, error) {
		return &spyProcessor{}, nil
	})
	Register("aa-unit", func(_ Config, _ string) (Processor, error) {
		return &spyProcessor{}, nil
	})

	names := RegisteredProcessors()
	if !sort.StringsAreSorted(names) {
		t.Fatalf("RegisteredProcessors() = %v, want sorted", names)
	}

	for _, want := range []string{"aa-unit", "zz-unit"} {
		if !contains(names, want) {
			t.Fatalf("RegisteredProcessors() = %v, want %q", names, want)
		}
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

func TestClientReturnsGaugeAndDistributionErrors(t *testing.T) {
	c := &Client{processors: []Processor{&spyProcessor{err: errors.New("processor failed")}}}

	if err := c.Gauge("active_connections", 1, nil); err == nil {
		t.Fatal("expected gauge error")
	}

	if err := c.Distribution("wait_seconds", 1, nil); err == nil {
		t.Fatal("expected distribution error")
	}
}

func TestCloseAggregatesProcessorErrors(t *testing.T) {
	firstErr := errors.New("first close failed")
	secondErr := errors.New("second close failed")
	first := &spyProcessor{err: firstErr}
	second := &spyProcessor{err: secondErr}
	c := &Client{processors: []Processor{nil, first, second}}

	err := c.Close()
	if !errors.Is(err, firstErr) {
		t.Fatalf("Close() error = %v, want %v", err, firstErr)
	}

	if !errors.Is(err, secondErr) {
		t.Fatalf("Close() error = %v, want %v", err, secondErr)
	}

	first.requireCall(t, "close")
	second.requireCall(t, "close")
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

func TestNormalizeTags(t *testing.T) {
	got := TagSet([]string{"status:200", "method:GET"})
	if got != "method:GET,status:200" {
		t.Fatalf("TagSet() = %q", got)
	}
}

func TestNormalizeTagsReturnsSortedCopy(t *testing.T) {
	if got := NormalizeTags(nil); got != nil {
		t.Fatalf("NormalizeTags(nil) = %v, want nil", got)
	}

	input := []string{"status:200", "method:GET"}
	got := NormalizeTags(input)
	want := []string{"method:GET", "status:200"}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NormalizeTags() = %v, want %v", got, want)
		}
	}

	got[0] = "mutated:true"

	if input[0] != "status:200" {
		t.Fatalf("NormalizeTags() returned alias; input = %v", input)
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

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
