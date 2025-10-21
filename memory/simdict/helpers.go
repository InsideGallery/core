package simdict

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"

	"github.com/sugarme/tokenizer/normalizer"

	"github.com/InsideGallery/core/utils"
)

func shingle(text string, k int) map[string]struct{} {
	shingles := make(map[string]struct{})

	text = strings.ToLower(strings.TrimSpace(text))
	for i := 0; i < len(text)-k+1; i++ {
		shingles[text[i:i+k]] = struct{}{}
	}

	return shingles
}

func createSignature(shingles map[string]struct{}, numHashes int) []uint32 {
	signature := make([]uint32, numHashes)
	for i := range signature {
		signature[i] = math.MaxUint32
	}

	for s := range shingles {
		for i := 0; i < numHashes; i++ {
			h := fnv.New32()
			fmt.Fprintf(h, "%d", i) // Seed
			h.Write([]byte(s))

			hashValue := h.Sum32()
			if hashValue < signature[i] {
				signature[i] = hashValue
			}
		}
	}

	return signature
}

func jaccardFromSignatures(sig1, sig2 []uint32) float64 {
	if len(sig1) != len(sig2) || len(sig1) == 0 {
		return 0.0
	}

	intersect := 0

	for i := 0; i < len(sig1); i++ {
		if sig1[i] == sig2[i] {
			intersect++
		}
	}

	return float64(intersect) / float64(len(sig1))
}

func Normalize(str string) (string, error) {
	n := normalizer.NewBertNormalizer(true, true, true, true)

	res, err := n.Normalize(normalizer.NewNormalizedFrom(utils.NFDLowerString(utils.SanitizeEmail(str))))
	if err != nil {
		return "", err
	}

	return res.GetNormalized(), nil
}
