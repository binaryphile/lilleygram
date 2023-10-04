package slice

func Map[T, R any](fn func(T) R, ts []T) []R {
	rs := make([]R, len(ts))

	for i, t := range ts {
		rs[i] = fn(t)
	}

	return rs
}
