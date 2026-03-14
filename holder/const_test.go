package holder_test

import (
	"testing"

	holder "github.com/holyheld/gointernals/holder"
)

func TestConstHolder(t *testing.T) {
	t.Parallel()

	t.Run("should return the same after acquiring a hold of original one", func(t *testing.T) {
		t.Parallel()

		original := 1

		h := holder.ConstHolder(original)
		if recv := h.Get(); recv != original {
			t.Errorf("Initial pass: Get() = %d, want %d", recv, original)
		}

		// Intentionally change value to mimic struct value being changed
		//nolint:ineffassign,wastedassign
		original = 2

		if recv := h.Get(); recv != 1 {
			t.Errorf("Second pass: Get() = %d, want %d", recv, 1)
		}
	})
}
