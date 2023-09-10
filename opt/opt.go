package opt

type Value[T any] struct {
	value T
	ok    bool
}

func Of[T any](value T, ok bool) (_ Value[T]) {
	if ok {
		return Value[T]{
			value: value,
			ok:    true,
		}
	}

	return
}

func OfNonZero[T comparable](value T) (zero Value[T]) {
	return Of(value, value != zero.value)
}
