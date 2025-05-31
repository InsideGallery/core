package hll

import (
	"github.com/segmentio/go-hll"
	"github.com/twmb/murmur3"
)

const (
	DefaultLog2m    = 31
	DefaultRegwidth = 8
)

var DefaultSettings = hll.Settings{
	Log2m:             DefaultLog2m,
	Regwidth:          DefaultRegwidth,
	ExplicitThreshold: hll.AutoExplicitThreshold,
	SparseEnabled:     true,
}

type HyperLogLog struct {
	HLL hll.Hll
}

func New() (*HyperLogLog, error) {
	h, err := hll.NewHll(DefaultSettings)
	if err != nil {
		return nil, err
	}

	return &HyperLogLog{
		HLL: h,
	}, nil
}

func FromBytes(raw []byte) (*HyperLogLog, error) {
	h, err := hll.FromBytes(raw)
	if err != nil {
		return nil, err
	}

	return &HyperLogLog{
		HLL: h,
	}, nil
}

func (h *HyperLogLog) Union(h2 *HyperLogLog) (*HyperLogLog, error) {
	hc, err := hll.NewHll(DefaultSettings)
	if err != nil {
		return nil, err
	}

	hc.Union(h.HLL)
	hc.Union(h2.HLL)

	return &HyperLogLog{HLL: hc}, nil
}

func (h *HyperLogLog) UnionCount(h2 *HyperLogLog) (uint64, error) {
	hc, err := h.Union(h2)
	if err != nil {
		return 0, err
	}

	return hc.Count(), nil
}

func (h *HyperLogLog) Add(data []byte) {
	h.HLL.AddRaw(murmur3.Sum64(data))
}

func (h *HyperLogLog) AddAny(data any) error {
	v, err := GetBytes(data)
	if err != nil {
		return err
	}

	h.HLL.AddRaw(murmur3.Sum64(v))

	return nil
}

func (h *HyperLogLog) ToBytes() []byte {
	return h.HLL.ToBytes()
}

func (h *HyperLogLog) Count() uint64 {
	return h.HLL.Cardinality()
}

func (h *HyperLogLog) IntersectionCount(h2 *HyperLogLog) (uint64, error) {
	hc, err := hll.NewHll(DefaultSettings)
	if err != nil {
		return 0, err
	}

	hc.Union(h.HLL)
	hc.Union(h2.HLL)

	return h.HLL.Cardinality() + h2.HLL.Cardinality() - hc.Cardinality(), nil
}
