package opt

type Value[T any] struct {
	ok bool
	v  T
}

func Map[T, R any](f func(T) R) func(Value[T]) Value[R] {
	return func(value Value[T]) (_ Value[R]) {
		if value.ok {
			return OfOk(f(value.v))
		}

		return
	}
}

func Of[T any](value T, ok bool) (_ Value[T]) {
	if ok {
		return Value[T]{
			ok: true,
			v:  value,
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
	return Of(value, value != zero.v)
}

func OfOk[T any](value T) Value[T] {
	return Value[T]{
		ok: true,
		v:  value,
	}
}

func OkOrNot[T, R any](value Value[T], ifOk, ifNot R) R {
	if value.ok {
		return ifOk
	}

	return ifNot
}

func (x Value[T]) Or(other T) T {
	if x.ok {
		return x.v
	}

	return other
}

func (x Value[T]) OrZero() (_ T) {
	if x.ok {
		return x.v
	}

	return
}

func (x Value[T]) IsOk() bool {
	return x.ok
}
