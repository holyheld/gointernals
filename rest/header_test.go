package rest_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/holyheld/rest"
)

func TestCopyHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dst    *http.Header
		src    http.Header
		expect http.Header
	}{
		{
			name: "Should copy headers to dst (no conflicts, single entry per key)",
			dst:  &http.Header{},
			src: http.Header{
				"Accept":       []string{"text/html"},
				"Content-Type": []string{"application/json"},
			},
			expect: http.Header{
				"Accept":       []string{"text/html"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name: "Should copy headers to dst (conflict: duplicate key, single entry per key)",
			dst:  &http.Header{"Accept": []string{"application/binary"}},
			src: http.Header{
				"Accept":       []string{"text/html"},
				"Content-Type": []string{"application/json"},
			},
			expect: http.Header{
				"Accept":       []string{"application/binary", "text/html"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name: "Should copy headers to dst with (no conflicts, multiple entries per key)",
			dst:  &http.Header{},
			src: http.Header{
				"Accept":       []string{"text/html", "application/json"},
				"Content-Type": []string{"application/json"},
			},
			expect: http.Header{
				"Accept":       []string{"text/html", "application/json"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name: "Should copy headers to dst with merge (duplicate key, multiple entries per key)",
			dst:  &http.Header{"Accept": []string{"text/markdown"}},
			src: http.Header{"Accept": []string{
				"text/html",
				"application/json",
			}, "Content-Type": []string{"application/json"}},
			expect: http.Header{
				"Accept":       []string{"text/markdown", "text/html", "application/json"},
				"Content-Type": []string{"application/json"},
			},
		},
		{
			name: "Should copy headers to dst with merge (duplicate key, entry, single entry per key)",
			dst:  &http.Header{"Accept": []string{"text/html"}},
			src: http.Header{
				"Accept":       []string{"text/html"},
				"Content-Type": []string{"application/json"},
			},
			expect: http.Header{
				"Accept":       []string{"text/html"},
				"Content-Type": []string{"application/json"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rest.CopyHeader(tt.dst, tt.src)

			err := compareHeader(t, tt.expect, *tt.dst)
			if err != nil {
				t.Errorf("compare failed: %s", err)
			}
		})
	}
}

func compareHeader(t *testing.T, want http.Header, got http.Header) error {
	t.Helper()

	if len(want) != len(got) {
		return fmt.Errorf("slices are different length: want %d got %d", len(want), len(got))
	}

	for k, wantGroup := range want {
		err := compareSlices(t, wantGroup, got.Values(k))
		if err != nil {
			return fmt.Errorf("want slices under key %s to be equal: %w", k, err)
		}
	}

	return nil
}

func compareSlices(t *testing.T, want []string, got []string) error {
	t.Helper()

	if len(want) != len(got) {
		return fmt.Errorf("slices are different length: want %d got %d", len(want), len(got))
	}

	for idx := range want {
		if want[idx] != got[idx] {
			return fmt.Errorf("idx %d want %s got %s", idx, want[idx], got[idx])
		}
	}

	return nil
}
