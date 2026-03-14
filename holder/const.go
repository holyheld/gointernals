package holder

type constHolder[T any] struct {
	value T
}

func (h *constHolder[T]) Get() T {
	return h.value
}

func ConstHolder[T any](value T) Holder[T] {
	return &constHolder[T]{value}
}
