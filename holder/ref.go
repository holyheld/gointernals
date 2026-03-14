package holder

import "github.com/holyheld/ref"

type refHolder[T any] struct {
	ref *T
}

func RefHolder[T any](ref *T) Holder[T] {
	return &refHolder[T]{ref}
}

func (h *refHolder[T]) Get() T {
	return ref.Unwrap(h.ref)
}
