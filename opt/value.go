package opt

type Value[T any] struct {
	ok bool
	v  T
}

func (v Value[T]) AndDo(fn func()) {
	if v.ok {
		fn()
	}
}

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

func (v Value[T]) Unpack() (_ T, _ bool) {
	if v.ok {
		return v.v, true
	}

	return
}
