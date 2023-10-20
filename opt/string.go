package opt

type String struct {
	Value[string]
}

func (s String) Filter(fn func(string) bool) (_ String) {
	if s.ok && fn(s.v) {
		return s
	}

	return
}
