package spatial

func Map[T, V any](t []T, f func(T) V) []V {
	var v []V
	for _, tt := range t {
		v = append(v, f(tt))
	}

	return v
}
