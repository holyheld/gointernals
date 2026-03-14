package typeutil

import (
	"iter"
	"slices"
)

// DeduplicateSeq accepts iterator sequence, and returns unique values sequence.
func DeduplicateSeq[T comparable](seq iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for value := range seq {
			_, present := seen[value]
			if present {
				continue
			}

			if !yield(value) {
				break
			}

			seen[value] = struct{}{}
		}
	}
}

// ChanSeq accepts iterator sequence and returns the channel representation of it.
func ChanSeq[T any](seq iter.Seq[T]) <-chan T {
	c := make(chan T)

	go func() {
		defer close(c)

		for value := range seq {
			c <- value
		}
	}()

	return c
}

// Map accepts iterator sequence and map function, returns sequence of products.
func Map[T any, U any](seq iter.Seq[T], fn func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				break
			}
		}
	}
}

func CopySeq[T any](seq iter.Seq[T]) (iter.Seq[T], iter.Seq[T]) {
	values := slices.Collect(seq)

	return slices.Values(values), slices.Values(values)
}
