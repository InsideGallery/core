package throughput

import (
	"sync/atomic"
	"time"

	"github.com/InsideGallery/core/memory/orderedmap"
)

const Days30 = time.Hour * 24 * 30

type Storage interface {
	RPS(key string) uint64
	RPM(key string) uint64
	Add(key string, tier int)
	Incr(key string)
	Tier(key string) int
	Reset()
}

type MemoryStorage struct {
	counter  *orderedmap.OrderedMap[string, *atomic.Uint64]
	counterM *orderedmap.OrderedMap[string, *atomic.Uint64]
	date     *orderedmap.OrderedMap[string, time.Time]
	tier     *orderedmap.OrderedMap[string, int]
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		counter:  &orderedmap.OrderedMap[string, *atomic.Uint64]{},
		counterM: &orderedmap.OrderedMap[string, *atomic.Uint64]{},
		date:     &orderedmap.OrderedMap[string, time.Time]{},
		tier:     &orderedmap.OrderedMap[string, int]{},
	}
}

func (s *MemoryStorage) Add(key string, tier int) {
	s.tier.Add(key, tier)
	s.date.Add(key, time.Now())
}

func (s *MemoryStorage) RPM(key string) uint64 {
	if !s.counterM.Exists(key) {
		s.counterM.Add(key, &atomic.Uint64{})
	}

	at := s.counterM.Get(key)

	return at.Load()
}

func (s *MemoryStorage) RPS(key string) uint64 {
	if !s.counter.Exists(key) {
		s.counter.Add(key, &atomic.Uint64{})
	}

	at := s.counter.Get(key)

	return at.Load()
}

func (s *MemoryStorage) Tier(key string) int {
	tier := Tier0
	if s.tier.Exists(key) {
		tier = s.tier.Get(key)
	}

	return tier
}

func (s *MemoryStorage) Incr(key string) {
	if !s.counter.Exists(key) {
		s.counter.Add(key, &atomic.Uint64{})
	}

	if !s.counterM.Exists(key) {
		s.counterM.Add(key, &atomic.Uint64{})
	}

	at := s.counter.Get(key)
	at.Add(1)

	at = s.counterM.Get(key)
	at.Add(1)
}

func (s *MemoryStorage) Reset() {
	for _, v := range s.counter.GetMap() {
		go func() {
			v.Store(0)
		}()
	}

	for k, v := range s.date.GetMap() {
		go func() {
			if v.Add(Days30).Before(time.Now()) {
				s.date.Add(k, time.Now())
				s.counterM.Add(k, &atomic.Uint64{})
			}
		}()
	}
}
