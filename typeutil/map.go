package typeutil

import (
	"cmp"
	"maps"
	"slices"
)

// Sorted returns map with key-value pairs sorted in order of keys.
func Sorted[K cmp.Ordered, V any](m map[K]V) map[K]V {
	sorted := make(map[K]V, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		sorted[k] = m[k]
	}

	return sorted
}
