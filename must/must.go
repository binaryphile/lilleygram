package must

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func Must1[T, R any](f func(T) (R, error)) func(T) R {
	return func(t T) R {
		r, err := f(t)
		if err != nil {
			panic(err)
		}

		return r
	}
}

func Must2[T, T2, R any](f func(T, T2) (R, error)) func(T, T2) R {
	return func(t T, t2 T2) R {
		r, err := f(t, t2)
		if err != nil {
			panic(err)
		}

		return r
	}
}
