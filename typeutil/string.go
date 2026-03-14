package typeutil

// TruncateString returns at most limit characters of provided string.
func TruncateString(s string, limit int) string {
	if limit > len(s) {
		return s
	}

	return s[:limit]
}
