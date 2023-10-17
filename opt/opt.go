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

func (v Value[T]) OrZeroAndDo(fn func()) (_ T) {
	if v.ok {
		return v.v
	}

	fn()

	return
}

func Apply[T, R any](fn func(T) R, v Value[T]) (_ Value[R]) {
	if v.ok {
		return Of(fn(v.v), true)
	}

	return
}

func Getenv(key string) Value[string] {
	return OfNonZero(os.Getenv(key))
}

func Of[T any](t T, ok bool) (_ Value[T]) {
	if ok {
		return OfOk(t)
	}

	return
}

func OfAssert[R, T any](t T) Value[R] {
	v, ok := any(t).(R)

	return Of(v, ok)
}

func OfFirst[T any](values []T) (_ Value[T]) {
	if len(values) > 0 {
		return OfOk(values[0])
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
