package passwordutil_test

import (
	"testing"

	"github.com/holyheld/passwordutil"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Under 26",
			length: 25,
		},
		{
			name:   "Exactly 26",
			length: 25,
		},
		{
			name:   "2 blocks of 26",
			length: 52,
		},
		{
			name:   "Uneven blocks of 26",
			length: 57,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := passwordutil.Generate(tt.length)
			if len(got) != tt.length {
				t.Errorf("Generate() = %v, want %v", got, tt.length)
			}
		})
	}
}
