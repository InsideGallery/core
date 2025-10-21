package simdict

import (
	"fmt"
	"hash/fnv"
)

type LSHIndex struct {
	bands       map[int]map[string][]string
	numBands    int
	rowsPerBand int
}

func NewLSHIndex(numBands, rowsPerBand int) *LSHIndex {
	return &LSHIndex{
		bands:       make(map[int]map[string][]string),
		numBands:    numBands,
		rowsPerBand: rowsPerBand,
	}
}

func (l *LSHIndex) Add(id string, signature []uint32) {
	for b := 0; b < l.numBands; b++ {
		start := b * l.rowsPerBand
		end := start + l.rowsPerBand
		bandSignature := signature[start:end]

		h := fnv.New64a()
		for _, val := range bandSignature {
			fmt.Fprintf(h, "%d", val)
		}

		bandHash := fmt.Sprintf("%x", h.Sum64())

		if _, ok := l.bands[b]; !ok {
			l.bands[b] = make(map[string][]string)
		}

		l.bands[b][bandHash] = append(l.bands[b][bandHash], id)
	}
}

func (l *LSHIndex) Query(signature []uint32) []string {
	candidates := make(map[string]struct{})

	for b := 0; b < l.numBands; b++ {
		start := b * l.rowsPerBand
		end := start + l.rowsPerBand
		bandSignature := signature[start:end]

		h := fnv.New64a()
		for _, val := range bandSignature {
			fmt.Fprintf(h, "%d", val)
		}

		bandHash := fmt.Sprintf("%x", h.Sum64())

		if bucket, ok := l.bands[b][bandHash]; ok {
			for _, id := range bucket {
				candidates[id] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(candidates))
	for id := range candidates {
		result = append(result, id)
	}

	return result
}
