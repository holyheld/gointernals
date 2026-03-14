package ref

//go:fix inline
func Take[T any](value T) *T {
	return new(value)
}
