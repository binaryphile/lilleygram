package sqlrepo

func ifThenElse[T any](cond bool, first, second T) T {
	if cond {
		return first
	}

	return second
}
