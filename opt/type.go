package opt

type Type[T, R any] struct {
	Value[T]
}

func NewType[T, R any](value Value[T]) Type[T, R] {
	return Type[T, R]{
		Value: value,
	}
}

func (t Type[T, R]) Filter(fn func(T) bool) (_ Type[T, R]) {
	if t.ok && fn(t.v) {
		return t
	}

	return
}

func (t Type[T, R]) Apply(fn func(T) R) (_ Value[R]) {
	if t.ok {
		return OfOk(fn(t.v))
	}

	return
}
