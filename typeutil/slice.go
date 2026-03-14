package typeutil

import "slices"

// SafeSlice returns always non-nil slice of T
//
// Useful to return in DTOs or functions which results are relied upon during marshaling.
func SafeSlice[T any](x []T) []T {
	if x == nil {
		return make([]T, 0)
	}

	return x
}

// ChanSlice creates a read-only channel with values from provided slice.
func ChanSlice[T any](vs []T) <-chan T {
	c := make(chan T, len(vs))
	defer close(c)

	for _, v := range vs {
		c <- v
	}

	return c
}

// Chunk returns a slice of chunks of original slice. All chunks except the last one
// are of size chunkSize.
func Chunk[T any](s []T, chunkSize int) [][]T {
	return slices.Collect(slices.Chunk(s, chunkSize))
}
