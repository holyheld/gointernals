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
//
// Keep in mind that the channel is unbuffered, so can lead to deadlocks if routine
// is unable to read from it due to external locks.
func ChanSeq[T any](seq iter.Seq[T]) <-chan T {
	return chanSeq(seq, 0)
}

// ChanSeqSized accepts iterator sequence and channel size, and returns
// the channel representation of iterator.
func ChanSeqSized[T any](seq iter.Seq[T], size int) <-chan T {
	return chanSeq(seq, size)
}

func chanSeq[T any](seq iter.Seq[T], size int) <-chan T {
	c := make(chan T, size)

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
