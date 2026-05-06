package ecs

import (
	"context"
	"math"
	"sync"
	"testing"
)

func TestNewBaseEntityWithID_Values(t *testing.T) {
	cases := []struct {
		name    string
		id      uint64
		wantID  uint64
		wantVer uint64
	}{
		{
			name:    "zero id",
			id:      0,
			wantID:  0,
			wantVer: 1,
		},
		{
			name:    "id one",
			id:      1,
			wantID:  1,
			wantVer: 1,
		},
		{
			name:    "large id",
			id:      999999999,
			wantID:  999999999,
			wantVer: 1,
		},
		{
			name:    "max uint64",
			id:      math.MaxUint64,
			wantID:  math.MaxUint64,
			wantVer: 1,
		},
		{
			name:    "mid range id",
			id:      math.MaxUint64 / 2,
			wantID:  math.MaxUint64 / 2,
			wantVer: 1,
		},
		{
			name:    "power of two",
			id:      1 << 32,
			wantID:  1 << 32,
			wantVer: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			if e.GetID() != tc.wantID {
				t.Fatalf("GetID() = %d, want %d", e.GetID(), tc.wantID)
			}
			if e.GetVersion() != tc.wantVer {
				t.Fatalf("GetVersion() = %d, want %d", e.GetVersion(), tc.wantVer)
			}
		})
	}
}

func TestNewBaseEntity_Sequential(t *testing.T) {
	cases := []struct {
		name  string
		count int
	}{
		{
			name:  "single entity",
			count: 1,
		},
		{
			name:  "two entities have sequential ids",
			count: 2,
		},
		{
			name:  "five entities have sequential ids",
			count: 5,
		},
		{
			name:  "ten entities have sequential ids",
			count: 10,
		},
		{
			name:  "twenty entities have sequential ids",
			count: 20,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entities := make([]*BaseEntity, tc.count)
			for i := 0; i < tc.count; i++ {
				entities[i] = NewBaseEntity()
			}
			for i := 1; i < tc.count; i++ {
				if entities[i].GetID() <= entities[i-1].GetID() {
					t.Fatalf("entity %d ID (%d) should be greater than entity %d ID (%d)",
						i, entities[i].GetID(), i-1, entities[i-1].GetID())
				}
			}
			for i := 0; i < tc.count; i++ {
				if entities[i].GetVersion() != 1 {
					t.Fatalf("entity %d version = %d, want 1", i, entities[i].GetVersion())
				}
			}
		})
	}
}

func TestGetID_WithNewBaseEntityWithID(t *testing.T) {
	cases := []struct {
		name   string
		id     uint64
		wantID uint64
	}{
		{
			name:   "zero",
			id:     0,
			wantID: 0,
		},
		{
			name:   "one",
			id:     1,
			wantID: 1,
		},
		{
			name:   "max uint64",
			id:     math.MaxUint64,
			wantID: math.MaxUint64,
		},
		{
			name:   "max uint32",
			id:     math.MaxUint32,
			wantID: math.MaxUint32,
		},
		{
			name:   "large value",
			id:     1<<63 - 1,
			wantID: 1<<63 - 1,
		},
		{
			name:   "small value 42",
			id:     42,
			wantID: 42,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			if got := e.GetID(); got != tc.wantID {
				t.Fatalf("GetID() = %d, want %d", got, tc.wantID)
			}
		})
	}
}

func TestSetID_UpdatesEntityID(t *testing.T) {
	cases := []struct {
		name      string
		initialID uint64
		newID     uint64
		wantID    uint64
	}{
		{
			name:      "set to zero",
			initialID: 100,
			newID:     0,
			wantID:    0,
		},
		{
			name:      "set to same value",
			initialID: 50,
			newID:     50,
			wantID:    50,
		},
		{
			name:      "set to larger value",
			initialID: 10,
			newID:     1000,
			wantID:    1000,
		},
		{
			name:      "set to smaller value",
			initialID: 1000,
			newID:     5,
			wantID:    5,
		},
		{
			name:      "set zero to one",
			initialID: 0,
			newID:     1,
			wantID:    1,
		},
		{
			name:      "set to moderate large value",
			initialID: 0,
			newID:     1 << 48,
			wantID:    1 << 48,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.initialID)
			e.SetID(tc.newID)
			if got := e.GetID(); got != tc.wantID {
				t.Fatalf("after SetID(%d), GetID() = %d, want %d", tc.newID, got, tc.wantID)
			}
		})
	}
}

func TestSetID_AdvancesStoreForNewEntities(t *testing.T) {
	cases := []struct {
		name  string
		setID uint64
	}{
		{
			name:  "set to moderate value",
			setID: 50_000_000,
		},
		{
			name:  "set to small value",
			setID: 100,
		},
		{
			name:  "set to power of two",
			setID: 1 << 20,
		},
		{
			name:  "set to large odd value",
			setID: 70_000_001,
		},
		{
			name:  "set to another large value",
			setID: 80_000_000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := NewBaseEntity()
			e := NewBaseEntityWithID(0)
			e.SetID(tc.setID)
			if got := e.GetID(); got != tc.setID {
				t.Fatalf("after SetID(%d), GetID() = %d", tc.setID, got)
			}
			after := NewBaseEntity()
			if after.GetID() <= before.GetID() {
				t.Fatalf("after SetID(%d), next NewBaseEntity ID %d should be > previous %d",
					tc.setID, after.GetID(), before.GetID())
			}
		})
	}
}

func TestGetVersion_InitialAndAfterSet(t *testing.T) {
	cases := []struct {
		name       string
		setVersion uint64
		useSet     bool
		wantVer    uint64
	}{
		{
			name:    "initial version is 1",
			useSet:  false,
			wantVer: 1,
		},
		{
			name:       "set to 100",
			useSet:     true,
			setVersion: 100,
			wantVer:    100,
		},
		{
			name:       "set to max uint64",
			useSet:     true,
			setVersion: math.MaxUint64,
			wantVer:    math.MaxUint64,
		},
		{
			name:       "set to one",
			useSet:     true,
			setVersion: 1,
			wantVer:    1,
		},
		{
			name:       "set to zero",
			useSet:     true,
			setVersion: 0,
			wantVer:    0,
		},
		{
			name:       "set to large value",
			useSet:     true,
			setVersion: 1<<48 + 3,
			wantVer:    1<<48 + 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			if tc.useSet {
				e.SetVersion(tc.setVersion)
			}
			if got := e.GetVersion(); got != tc.wantVer {
				t.Fatalf("GetVersion() = %d, want %d", got, tc.wantVer)
			}
		})
	}
}

func TestUpVersion_FromDefault(t *testing.T) {
	cases := []struct {
		name    string
		upCount int
		wantVer uint64
	}{
		{
			name:    "single increment from default",
			upCount: 1,
			wantVer: 2,
		},
		{
			name:    "five increments from default",
			upCount: 5,
			wantVer: 6,
		},
		{
			name:    "zero increments stays at default",
			upCount: 0,
			wantVer: 1,
		},
		{
			name:    "one hundred increments",
			upCount: 100,
			wantVer: 101,
		},
		{
			name:    "two increments",
			upCount: 2,
			wantVer: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			for i := 0; i < tc.upCount; i++ {
				e.UpVersion()
			}
			if got := e.GetVersion(); got != tc.wantVer {
				t.Fatalf("after %d UpVersion() calls, GetVersion() = %d, want %d",
					tc.upCount, got, tc.wantVer)
			}
		})
	}
}

func TestUpVersion_FromSetVersion(t *testing.T) {
	cases := []struct {
		name       string
		initialVer uint64
		upCount    int
		wantVer    uint64
	}{
		{
			name:       "set 50 then increment once",
			initialVer: 50,
			upCount:    1,
			wantVer:    51,
		},
		{
			name:       "set 0 then increment once",
			initialVer: 0,
			upCount:    1,
			wantVer:    1,
		},
		{
			name:       "set 10 then increment 5",
			initialVer: 10,
			upCount:    5,
			wantVer:    15,
		},
		{
			name:       "set 100 then no increment",
			initialVer: 100,
			upCount:    0,
			wantVer:    100,
		},
		{
			name:       "set max-1 then increment once reaches max",
			initialVer: math.MaxUint64 - 1,
			upCount:    1,
			wantVer:    math.MaxUint64,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			e.SetVersion(tc.initialVer)
			for i := 0; i < tc.upCount; i++ {
				e.UpVersion()
			}
			if got := e.GetVersion(); got != tc.wantVer {
				t.Fatalf("SetVersion(%d) then %d UpVersion() = %d, want %d",
					tc.initialVer, tc.upCount, got, tc.wantVer)
			}
		})
	}
}

func TestSetVersion_Values(t *testing.T) {
	cases := []struct {
		name    string
		version uint64
		want    uint64
	}{
		{
			name:    "set to zero",
			version: 0,
			want:    0,
		},
		{
			name:    "set to one",
			version: 1,
			want:    1,
		},
		{
			name:    "set to max uint64",
			version: math.MaxUint64,
			want:    math.MaxUint64,
		},
		{
			name:    "set to large value",
			version: 1<<62 + 99,
			want:    1<<62 + 99,
		},
		{
			name:    "set to 42",
			version: 42,
			want:    42,
		},
		{
			name:    "set to max uint32",
			version: math.MaxUint32,
			want:    math.MaxUint32,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			e.SetVersion(tc.version)
			if got := e.GetVersion(); got != tc.want {
				t.Fatalf("SetVersion(%d), GetVersion() = %d, want %d", tc.version, got, tc.want)
			}
		})
	}
}

func TestSetVersion_Overwrites(t *testing.T) {
	cases := []struct {
		name   string
		first  uint64
		second uint64
		want   uint64
	}{
		{
			name:   "overwrite high with low",
			first:  1000,
			second: 1,
			want:   1,
		},
		{
			name:   "overwrite low with high",
			first:  1,
			second: 1000,
			want:   1000,
		},
		{
			name:   "overwrite with same",
			first:  42,
			second: 42,
			want:   42,
		},
		{
			name:   "overwrite with zero",
			first:  999,
			second: 0,
			want:   0,
		},
		{
			name:   "overwrite zero with max",
			first:  0,
			second: math.MaxUint64,
			want:   math.MaxUint64,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			e.SetVersion(tc.first)
			e.SetVersion(tc.second)
			if got := e.GetVersion(); got != tc.want {
				t.Fatalf("after SetVersion(%d) then SetVersion(%d), GetVersion() = %d, want %d",
					tc.first, tc.second, got, tc.want)
			}
		})
	}
}

func TestUpVersion_Concurrent(t *testing.T) {
	cases := []struct {
		name       string
		goroutines int
		perG       int
		wantVer    uint64
	}{
		{
			name:       "2 goroutines 50 each",
			goroutines: 2,
			perG:       50,
			wantVer:    101,
		},
		{
			name:       "10 goroutines 10 each",
			goroutines: 10,
			perG:       10,
			wantVer:    101,
		},
		{
			name:       "1 goroutine 100 calls",
			goroutines: 1,
			perG:       100,
			wantVer:    101,
		},
		{
			name:       "5 goroutines 20 each",
			goroutines: 5,
			perG:       20,
			wantVer:    101,
		},
		{
			name:       "100 goroutines 1 each",
			goroutines: 100,
			perG:       1,
			wantVer:    101,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			var wg sync.WaitGroup
			wg.Add(tc.goroutines)
			for g := 0; g < tc.goroutines; g++ {
				go func() {
					defer wg.Done()
					for i := 0; i < tc.perG; i++ {
						e.UpVersion()
					}
				}()
			}
			wg.Wait()
			if got := e.GetVersion(); got != tc.wantVer {
				t.Fatalf("concurrent UpVersion: GetVersion() = %d, want %d", got, tc.wantVer)
			}
		})
	}
}

func TestEntityInterface_Satisfaction(t *testing.T) {
	cases := []struct {
		name string
		id   uint64
	}{
		{
			name: "zero id implements Entity",
			id:   0,
		},
		{
			name: "nonzero id implements Entity",
			id:   42,
		},
		{
			name: "large id implements Entity",
			id:   math.MaxUint64,
		},
		{
			name: "power of two implements Entity",
			id:   1 << 16,
		},
		{
			name: "one implements Entity",
			id:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			var iface Entity = e
			if iface.GetID() != tc.id {
				t.Fatalf("Entity interface GetID() = %d, want %d", iface.GetID(), tc.id)
			}
		})
	}
}

func TestVersionableInterface_Satisfaction(t *testing.T) {
	cases := []struct {
		name    string
		upCount int
		wantVer uint64
	}{
		{
			name:    "no increments",
			upCount: 0,
			wantVer: 1,
		},
		{
			name:    "one increment",
			upCount: 1,
			wantVer: 2,
		},
		{
			name:    "three increments",
			upCount: 3,
			wantVer: 4,
		},
		{
			name:    "ten increments",
			upCount: 10,
			wantVer: 11,
		},
		{
			name:    "fifty increments",
			upCount: 50,
			wantVer: 51,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			var iface Versionable = e
			for i := 0; i < tc.upCount; i++ {
				iface.UpVersion()
			}
			if got := iface.GetVersion(); got != tc.wantVer {
				t.Fatalf("Versionable GetVersion() = %d, want %d", got, tc.wantVer)
			}
		})
	}
}

type mockComponent struct {
	called bool
	err    error
}

func (m *mockComponent) Update(_ context.Context) error {
	m.called = true
	return m.err
}

func TestComponentInterface_Update(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCalled bool
		wantErr    bool
	}{
		{
			name:       "nil error component",
			err:        nil,
			wantCalled: true,
			wantErr:    false,
		},
		{
			name:       "deadline exceeded error",
			err:        context.DeadlineExceeded,
			wantCalled: true,
			wantErr:    true,
		},
		{
			name:       "canceled context error",
			err:        context.Canceled,
			wantCalled: true,
			wantErr:    true,
		},
		{
			name:       "success update",
			err:        nil,
			wantCalled: true,
			wantErr:    false,
		},
		{
			name:       "another error type",
			err:        context.DeadlineExceeded,
			wantCalled: true,
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockComponent{err: tc.err}
			var c Component = m
			err := c.Update(context.Background())
			if m.called != tc.wantCalled {
				t.Fatalf("called = %v, want %v", m.called, tc.wantCalled)
			}
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

type mockSystem struct {
	called bool
	err    error
}

func (m *mockSystem) Update(_ context.Context) error {
	m.called = true
	return m.err
}

func TestSystemInterface_Update(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCalled bool
		wantErr    bool
	}{
		{
			name:       "nil error system",
			err:        nil,
			wantCalled: true,
			wantErr:    false,
		},
		{
			name:       "deadline exceeded error",
			err:        context.DeadlineExceeded,
			wantCalled: true,
			wantErr:    true,
		},
		{
			name:       "canceled context error",
			err:        context.Canceled,
			wantCalled: true,
			wantErr:    true,
		},
		{
			name:       "success system call",
			err:        nil,
			wantCalled: true,
			wantErr:    false,
		},
		{
			name:       "another deadline exceeded",
			err:        context.DeadlineExceeded,
			wantCalled: true,
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockSystem{err: tc.err}
			var s System = m
			err := s.Update(context.Background())
			if m.called != tc.wantCalled {
				t.Fatalf("called = %v, want %v", m.called, tc.wantCalled)
			}
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewBaseEntity_IDIsPositive(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "first call"},
		{name: "second call"},
		{name: "third call"},
		{name: "fourth call"},
		{name: "fifth call"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntity()
			if e.GetID() == 0 {
				t.Fatal("NewBaseEntity should return non-zero ID")
			}
		})
	}
}

func TestNewBaseEntity_VersionAlwaysOne(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "check 1"},
		{name: "check 2"},
		{name: "check 3"},
		{name: "check 4"},
		{name: "check 5"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntity()
			if e.GetVersion() != 1 {
				t.Fatalf("NewBaseEntity version = %d, want 1", e.GetVersion())
			}
		})
	}
}

func TestSetID_DoesNotAffectVersion(t *testing.T) {
	cases := []struct {
		name  string
		newID uint64
	}{
		{name: "set id to 0", newID: 0},
		{name: "set id to 100", newID: 100},
		{name: "set id to max uint32", newID: math.MaxUint32},
		{name: "set id to 500", newID: 500},
		{name: "set id to 1", newID: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(50)
			e.UpVersion()
			ver := e.GetVersion()
			e.SetID(tc.newID)
			if e.GetVersion() != ver {
				t.Fatalf("SetID changed version from %d to %d", ver, e.GetVersion())
			}
		})
	}
}

func TestSetVersion_DoesNotAffectID(t *testing.T) {
	cases := []struct {
		name    string
		id      uint64
		version uint64
	}{
		{name: "id 10 version 0", id: 10, version: 0},
		{name: "id 10 version 100", id: 10, version: 100},
		{name: "id 10 version max", id: 10, version: math.MaxUint64},
		{name: "id 0 version 50", id: 0, version: 50},
		{name: "id 999 version 1", id: 999, version: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			e.SetVersion(tc.version)
			if e.GetID() != tc.id {
				t.Fatalf("SetVersion changed ID from %d to %d", tc.id, e.GetID())
			}
		})
	}
}

func TestMultipleEntities_Independent(t *testing.T) {
	cases := []struct {
		name    string
		id1     uint64
		id2     uint64
		ver1Set uint64
		ver2Set uint64
	}{
		{name: "different ids and versions", id1: 1, id2: 2, ver1Set: 10, ver2Set: 20},
		{name: "same id different versions", id1: 5, id2: 5, ver1Set: 100, ver2Set: 200},
		{name: "zero ids zero versions", id1: 0, id2: 0, ver1Set: 0, ver2Set: 0},
		{name: "adjacent ids", id1: 99, id2: 100, ver1Set: 5, ver2Set: 6},
		{name: "large ids", id1: 1 << 40, id2: 1 << 41, ver1Set: 77, ver2Set: 88},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e1 := NewBaseEntityWithID(tc.id1)
			e2 := NewBaseEntityWithID(tc.id2)
			e1.SetVersion(tc.ver1Set)
			e2.SetVersion(tc.ver2Set)
			if e1.GetVersion() != tc.ver1Set {
				t.Fatalf("e1 version = %d, want %d", e1.GetVersion(), tc.ver1Set)
			}
			if e2.GetVersion() != tc.ver2Set {
				t.Fatalf("e2 version = %d, want %d", e2.GetVersion(), tc.ver2Set)
			}
			if e1.GetID() != tc.id1 {
				t.Fatalf("e1 id = %d, want %d", e1.GetID(), tc.id1)
			}
			if e2.GetID() != tc.id2 {
				t.Fatalf("e2 id = %d, want %d", e2.GetID(), tc.id2)
			}
		})
	}
}

func TestConcurrentGetSetVersion(t *testing.T) {
	cases := []struct {
		name       string
		goroutines int
		iterations int
	}{
		{name: "2 goroutines 100 iterations", goroutines: 2, iterations: 100},
		{name: "4 goroutines 50 iterations", goroutines: 4, iterations: 50},
		{name: "10 goroutines 10 iterations", goroutines: 10, iterations: 10},
		{name: "1 goroutine 500 iterations", goroutines: 1, iterations: 500},
		{name: "50 goroutines 2 iterations", goroutines: 50, iterations: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			var wg sync.WaitGroup
			wg.Add(tc.goroutines * 2)
			for g := 0; g < tc.goroutines; g++ {
				go func() {
					defer wg.Done()
					for i := 0; i < tc.iterations; i++ {
						e.SetVersion(uint64(i))
					}
				}()
				go func() {
					defer wg.Done()
					for i := 0; i < tc.iterations; i++ {
						_ = e.GetVersion()
					}
				}()
			}
			wg.Wait()
		})
	}
}

func TestNewBaseEntity_ConcurrentUniqueIDs(t *testing.T) {
	cases := []struct {
		name       string
		goroutines int
		perG       int
	}{
		{name: "2 goroutines 10 each", goroutines: 2, perG: 10},
		{name: "5 goroutines 5 each", goroutines: 5, perG: 5},
		{name: "10 goroutines 1 each", goroutines: 10, perG: 1},
		{name: "1 goroutine 20 entities", goroutines: 1, perG: 20},
		{name: "4 goroutines 25 each", goroutines: 4, perG: 25},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			total := tc.goroutines * tc.perG
			ids := make(chan uint64, total)
			var wg sync.WaitGroup
			wg.Add(tc.goroutines)
			for g := 0; g < tc.goroutines; g++ {
				go func() {
					defer wg.Done()
					for i := 0; i < tc.perG; i++ {
						e := NewBaseEntity()
						ids <- e.GetID()
					}
				}()
			}
			wg.Wait()
			close(ids)

			seen := make(map[uint64]bool)
			for id := range ids {
				if seen[id] {
					t.Fatalf("duplicate ID: %d", id)
				}
				seen[id] = true
			}
			if len(seen) != total {
				t.Fatalf("expected %d unique IDs, got %d", total, len(seen))
			}
		})
	}
}

func TestVersionOverflow_Wrap(t *testing.T) {
	cases := []struct {
		name    string
		initial uint64
		ups     int
		want    uint64
	}{
		{name: "max wraps to 0", initial: math.MaxUint64, ups: 1, want: 0},
		{name: "max-1 goes to max", initial: math.MaxUint64 - 1, ups: 1, want: math.MaxUint64},
		{name: "max wraps twice to 1", initial: math.MaxUint64, ups: 2, want: 1},
		{name: "max-2 goes to max", initial: math.MaxUint64 - 2, ups: 2, want: math.MaxUint64},
		{name: "max wraps three times to 2", initial: math.MaxUint64, ups: 3, want: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			e.SetVersion(tc.initial)
			for i := 0; i < tc.ups; i++ {
				e.UpVersion()
			}
			if got := e.GetVersion(); got != tc.want {
				t.Fatalf("version overflow: got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNewBaseEntityWithID_DoesNotAffectGlobalCounter(t *testing.T) {
	cases := []struct {
		name string
		id   uint64
	}{
		{name: "id zero", id: 0},
		{name: "id one", id: 1},
		{name: "id max uint32", id: math.MaxUint32},
		{name: "id 42", id: 42},
		{name: "id 1000", id: 1000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := NewBaseEntity()
			_ = NewBaseEntityWithID(tc.id)
			after := NewBaseEntity()
			if after.GetID() != before.GetID()+1 {
				t.Fatalf("NewBaseEntityWithID(%d) affected global counter: before=%d, after=%d",
					tc.id, before.GetID(), after.GetID())
			}
		})
	}
}

func TestSetID_MultipleCallsOnSameEntity(t *testing.T) {
	cases := []struct {
		name string
		ids  []uint64
		want uint64
	}{
		{name: "ascending ids", ids: []uint64{1, 2, 3, 4, 5}, want: 5},
		{name: "descending ids", ids: []uint64{5, 4, 3, 2, 1}, want: 1},
		{name: "same id repeated", ids: []uint64{7, 7, 7, 7, 7}, want: 7},
		{name: "alternating ids", ids: []uint64{1, 100, 1, 100, 1}, want: 1},
		{name: "single set", ids: []uint64{42}, want: 42},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			for _, id := range tc.ids {
				e.SetID(id)
			}
			if got := e.GetID(); got != tc.want {
				t.Fatalf("after multiple SetID calls, GetID() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestUpVersion_ThenSetVersion_ThenUpVersion(t *testing.T) {
	cases := []struct {
		name      string
		ups1      int
		setTo     uint64
		ups2      int
		wantFinal uint64
	}{
		{name: "up1 set0 up1", ups1: 1, setTo: 0, ups2: 1, wantFinal: 1},
		{name: "up3 set10 up2", ups1: 3, setTo: 10, ups2: 2, wantFinal: 12},
		{name: "up0 set100 up0", ups1: 0, setTo: 100, ups2: 0, wantFinal: 100},
		{name: "up5 set5 up5", ups1: 5, setTo: 5, ups2: 5, wantFinal: 10},
		{name: "up10 set1 up1", ups1: 10, setTo: 1, ups2: 1, wantFinal: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(0)
			for i := 0; i < tc.ups1; i++ {
				e.UpVersion()
			}
			e.SetVersion(tc.setTo)
			for i := 0; i < tc.ups2; i++ {
				e.UpVersion()
			}
			if got := e.GetVersion(); got != tc.wantFinal {
				t.Fatalf("got version %d, want %d", got, tc.wantFinal)
			}
		})
	}
}

func TestNewBaseEntityWithID_VersionIndependentOfID(t *testing.T) {
	cases := []struct {
		name string
		id   uint64
	}{
		{name: "id 0", id: 0},
		{name: "id 1", id: 1},
		{name: "id max uint64", id: math.MaxUint64},
		{name: "id max uint32", id: math.MaxUint32},
		{name: "id 12345", id: 12345},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			if e.GetVersion() != 1 {
				t.Fatalf("NewBaseEntityWithID(%d) version = %d, want 1", tc.id, e.GetVersion())
			}
		})
	}
}

func TestBothInterfaces_OnSameEntity(t *testing.T) {
	cases := []struct {
		name    string
		id      uint64
		upCount int
	}{
		{name: "id 0 up 0", id: 0, upCount: 0},
		{name: "id 100 up 5", id: 100, upCount: 5},
		{name: "id 1 up 1", id: 1, upCount: 1},
		{name: "id max uint32 up 10", id: math.MaxUint32, upCount: 10},
		{name: "id 42 up 42", id: 42, upCount: 42},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewBaseEntityWithID(tc.id)
			var ent Entity = e
			var ver Versionable = e
			for i := 0; i < tc.upCount; i++ {
				ver.UpVersion()
			}
			if ent.GetID() != tc.id {
				t.Fatalf("Entity.GetID() = %d, want %d", ent.GetID(), tc.id)
			}
			wantVer := uint64(1) + uint64(tc.upCount)
			if ver.GetVersion() != wantVer {
				t.Fatalf("Versionable.GetVersion() = %d, want %d", ver.GetVersion(), wantVer)
			}
		})
	}
}
