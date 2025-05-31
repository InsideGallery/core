package mathutils

import (
	"crypto/rand"
	"math"
	"math/big"
)

// WeightIndex get key by probability chances
func WeightIndex(prob map[interface{}]uint64) interface{} {
	if len(prob) == 0 {
		return nil
	}

	indexes := map[int]interface{}{}
	probabilities := map[interface{}]uint64{}

	var count int
	var sumWeights uint64

	for key, v := range prob {
		sumWeights += v
		indexes[count] = key
		probabilities[key] = v
		count++
	}

	nBig, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(math.MaxUint64))
	maxUint64 := uint64(math.MaxUint64)
	r := uint64(float64(nBig.Uint64()) / float64(maxUint64) * float64(sumWeights))

	var prevp uint64

	for i := 0; i < count-1; i++ {
		if probabilities[indexes[i]] == 0 {
			continue
		}

		probabilities[indexes[i]] += prevp
		if r <= probabilities[indexes[i]] {
			return indexes[i]
		}

		prevp = probabilities[indexes[i]]
	}

	return indexes[count-1]
}
