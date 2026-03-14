package ref

func Take[T any](value T) *T {
	return &value
}
