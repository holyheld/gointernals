package slogutil_test

import (
	"errors"
	"testing"

	"github.com/holyheld/slogutil"
)

func TestRecover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    any
		want string
	}{
		{
			name: "string value",
			r:    generatePanic(t, "value"),
			want: "recover=value",
		},
		{
			name: "error value",
			r:    generatePanic(t, errors.New("value")),
			want: "recover=[message=value]",
		},
		{
			name: "other value",
			r:    generatePanic(t, 42),
			want: "recover=42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slogutil.Recover(tt.r)
			if got.String() != tt.want {
				t.Errorf("Recover() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generatePanic(t *testing.T, value any) (res any) {
	t.Helper()

	defer func() {
		res = recover()
	}()
	panic(value)
}
