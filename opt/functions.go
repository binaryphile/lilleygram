package opt

import "os"

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

func OfAssert[T any](a any) Value[T] {
	v, ok := a.(T)

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

func TypeOfNonZero[R any, T comparable](value T) (zero Type[T, R]) {
	return Type[T, R]{
		Value: Of(value, value != zero.v),
	}
}
