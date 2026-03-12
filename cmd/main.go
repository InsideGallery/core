package main

import (
	"fmt"
	"math/rand"
	"time"
)

type MinHash struct {
	numHashFunctions int
	hashFunctions    []func(string) int
}

// NewMinHash creates a new MinHash instance
func NewMinHash(numHashFunctions int) *MinHash {
	rand.Seed(time.Now().UnixNano())
	hashFunctions := make([]func(string) int, numHashFunctions)
	for i := 0; i < numHashFunctions; i++ {
		// Simple hash function using a random seed
		hashFunctions[i] = func(s string) int {
			return int(rand.Int63() + int64(i))
		}
	}
	return &MinHash{numHashFunctions, hashFunctions}
}

// ComputeMinHash computes the MinHash signature for a set of items
func (m *MinHash) ComputeMinHash(items []string) []int {
	signature := make([]int, m.numHashFunctions)
	for i := range signature {
		signature[i] = int(^uint(0) >> 1) // Max int
	}
	for _, item := range items {
		for i, hashFunc := range m.hashFunctions {
			hashValue := hashFunc(item)
			if hashValue < signature[i] {
				signature[i] = hashValue
			}
		}
	}
	return signature
}

func main() {
	minHash := NewMinHash(100)
	items := []string{"apple", "banana", "orange"}
	signature := minHash.ComputeMinHash(items)
	fmt.Println("MinHash Signature:", signature)
}
