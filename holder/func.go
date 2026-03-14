package holder

type funcHolder[T any] struct {
	f func() T
}

func (h *funcHolder[T]) Get() T {
	return h.f()
}

func FuncHolder[T any](f func() T) Holder[T] {
	return &funcHolder[T]{f}
}
