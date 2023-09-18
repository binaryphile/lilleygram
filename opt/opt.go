package opt

import "os"

type (
	Value[T any] struct {
		ok bool
		v  T
	}

	Int64 = Value[int64]

	String = Value[string]
)

func (v Value[T]) Or(ifNot T) T {
	if v.ok {
		return v.v
	}

	return ifNot
}

func (v Value[T]) OrZero() (_ T) {
	if v.ok {
		return v.v
	}

	return
}

func (v Value[T]) IsOk() bool {
	return v.ok
}

func Getenv(key string) Value[string] {
	return OfNonZero(os.Getenv(key))
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
