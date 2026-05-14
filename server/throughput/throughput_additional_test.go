package throughput

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryStorageOperations(t *testing.T) {
	storage := NewMemoryStorage()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "default counters and tier",
			run: func(t *testing.T) {
				t.Helper()

				if got := storage.RPS("unknown"); got != 0 {
					t.Fatalf("RPS = %d, want 0", got)
				}

				if got := storage.RPM("unknown"); got != 0 {
					t.Fatalf("RPM = %d, want 0", got)
				}

				if got := storage.Tier("unknown"); got != Tier0 {
					t.Fatalf("Tier = %d, want %d", got, Tier0)
				}
			},
		},
		{
			name: "add and increment",
			run: func(t *testing.T) {
				t.Helper()

				storage.Add("known", Tier2)
				storage.Incr("known")

				if got := storage.Tier("known"); got != Tier2 {
					t.Fatalf("Tier = %d, want %d", got, Tier2)
				}

				if got := storage.RPS("known"); got != 1 {
					t.Fatalf("RPS = %d, want 1", got)
				}

				if got := storage.RPM("known"); got != 1 {
					t.Fatalf("RPM = %d, want 1", got)
				}
			},
		},
		{
			name: "reset clears second counters and rotates old monthly counters",
			run: func(t *testing.T) {
				t.Helper()

				storage.counter.Add("old", &atomic.Uint64{})
				storage.counter.Get("old").Store(7)
				storage.counterM.Add("old", &atomic.Uint64{})
				storage.counterM.Get("old").Store(9)
				storage.date.Add("old", time.Now().Add(-Days30-time.Hour))

				storage.Reset()

				if got := storage.RPS("old"); got != 0 {
					t.Fatalf("RPS after reset = %d, want 0", got)
				}

				if got := storage.RPM("old"); got != 0 {
					t.Fatalf("RPM after reset = %d, want 0", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestThroughputValidateBranches(t *testing.T) {
	cases := []struct {
		name string
		fake *fakeStorage
		want bool
	}{
		{
			name: "rejects when rps limit reached",
			fake: &fakeStorage{tier: Tier0, rps: Tier0RPS},
		},
		{
			name: "rejects when rpm limit reached",
			fake: &fakeStorage{tier: Tier0, rpm: Tier0RPM},
		},
		{
			name: "increments accepted request",
			fake: &fakeStorage{tier: Tier1},
			want: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			throughput := New(context.Background(), test.fake)

			if got := throughput.Validate("client"); got != test.want {
				t.Fatalf("Validate() = %v, want %v", got, test.want)
			}

			if test.want && test.fake.increments != 1 {
				t.Fatalf("increments = %d, want 1", test.fake.increments)
			}
		})
	}
}

func TestThroughputLoopStopsWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		New(ctx, &fakeStorage{}).Loop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("loop did not stop")
	}
}

func TestTierLimits(t *testing.T) {
	cases := []struct {
		name    string
		tier    int
		wantRPS uint64
		wantRPM uint64
	}{
		{name: "tier zero", tier: Tier0, wantRPS: Tier0RPS, wantRPM: Tier0RPM},
		{name: "tier one", tier: Tier1, wantRPS: Tier1RPS, wantRPM: Tier1RPM},
		{name: "tier two", tier: Tier2, wantRPS: Tier2RPS, wantRPM: Tier2RPM},
		{name: "tier three", tier: Tier3, wantRPS: Tier3RPS, wantRPM: Tier3RPM},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := GetRPS(test.tier); got != test.wantRPS {
				t.Fatalf("RPS = %d, want %d", got, test.wantRPS)
			}

			if got := GetRPM(test.tier); got != test.wantRPM {
				t.Fatalf("RPM = %d, want %d", got, test.wantRPM)
			}
		})
	}
}

type fakeStorage struct {
	tier       int
	rps        uint64
	rpm        uint64
	increments int
	resets     int
}

func (s *fakeStorage) RPS(string) uint64 {
	return s.rps
}

func (s *fakeStorage) RPM(string) uint64 {
	return s.rpm
}

func (s *fakeStorage) Add(string, int) {}

func (s *fakeStorage) Incr(string) {
	s.increments++
}

func (s *fakeStorage) Tier(string) int {
	return s.tier
}

func (s *fakeStorage) Reset() {
	s.resets++
}
