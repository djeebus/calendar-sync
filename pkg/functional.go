package pkg

func ToMap[I any, K comparable](items []I, fn func(I) K) map[K]I {
	m := make(map[K]I)

	for _, i := range items {
		m[fn(i)] = i
	}

	return m
}

func ToSet[I any, K comparable](items []I, fn func(I) K) map[K]struct{} {
	m := make(map[K]struct{})

	for _, i := range items {
		m[fn(i)] = struct{}{}
	}

	return m
}

func Filter[I any](items []I, fn func(I) bool) []I {
	var result []I
	for _, i := range items {
		if fn(i) {
			result = append(result, i)
		}
	}
	return result
}
