package holder_test

import (
	"testing"

	holder "github.com/holyheld/gointernals/holder"
)

func TestRefHolder(t *testing.T) {
	t.Parallel()

	t.Run("should return the same after acquiring a hold of original one", func(t *testing.T) {
		t.Parallel()

		original := 1

		h := holder.RefHolder(&original)
		if recv := h.Get(); recv != original {
			t.Errorf("Initial pass: Get() = %d, want %d", recv, original)
		}

		// Intentionally change value to mimic struct value being changed

		original = 2

		if recv := h.Get(); recv != 2 {
			t.Errorf("Second pass: Get() = %d, want %d", recv, 2)
		}
	})

	t.Run(
		"should return different value after acquiring a hold of original one (ref of a struct by ref)",
		func(t *testing.T) {
			t.Parallel()

			str := &struct{ original int }{original: 1}

			h := holder.RefHolder(&str.original)
			if recv := h.Get(); recv != str.original {
				t.Errorf("Initial pass: Get() = %d, want %d", recv, str.original)
			}

			// Intentionally change value to mimic struct value being changed

			str.original = 2

			if recv := h.Get(); recv != 2 {
				t.Errorf("Second pass: Get() = %d, want %d", recv, 2)
			}
		},
	)
}
