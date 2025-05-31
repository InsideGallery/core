package utils

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
