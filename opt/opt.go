package opt

type Value[T any] struct {
	value T
	ok    bool
}

func Map[T, R any](f func(T) R, v Value[T]) (_ Value[R]) {
	if v.ok {
		return OfOk(f(v.value))
	}

	return
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

func OfIndex[K comparable, V any, M ~map[K]V](m M, k K) (_ Value[V]) {
	v, ok := m[k]

	if ok {
		return OfOk(v)
	}

	return
}

func OfNonZero[T comparable](value T) (zero Value[T]) {
	return Of(value, value != zero.value)
}

func OfOk[T any](value T) Value[T] {
	return Value[T]{
		ok:    true,
		value: value,
	}
}

func OkOrNot[T, R any](value Value[T], first, second R) R {
	if value.ok {
		return first
	}

	return second
}

func (x Value[T]) Or(other T) T {
	if x.ok {
		return x.value
	}

	return other
}

func (x Value[T]) OrZero() (_ T) {
	if x.ok {
		return x.value
	}

	return
}

func (x Value[T]) IsOk() bool {
	return x.ok
}
