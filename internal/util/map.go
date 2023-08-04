package util

func MapGetOrCreate[K comparable, V any](m map[K]V, k K, createFn func() V) V {
	v, ok := m[k]
	if !ok {
		v = createFn()
		m[k] = v
	}
	return v
}
