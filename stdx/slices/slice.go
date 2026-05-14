package slices

import "github.com/FrogoAI/set"

func BatchSlice[K any](size int, result []K) <-chan []K {
	if size <= 0 {
		size = 1
	}

	ch := make(chan []K, 1)
	count := len(result)
	parts := count / size

	go func() {
		defer close(ch)

		var begin, end int

		for i := 0; i < parts; i++ {
			begin = size * i

			end = begin + size
			ch <- result[begin:end]
		}

		if len(result[end:]) != 0 {
			ch <- result[end:]
		}
	}()

	return ch
}

func Shingle(text string, k int) set.GenericDataSet[string] {
	shingleSet := set.NewGenericDataSet[string]()

	if k <= 0 || len(text) == 0 {
		return shingleSet
	}

	if len(text) < k {
		shingleSet.Add(text)
		return shingleSet
	}

	for i := 0; i < len(text)-k+1; i++ {
		shingleSet.Add(text[i : i+k])
	}

	return shingleSet
}
