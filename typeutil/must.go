package typeutil

// Must is a wrapper to safely ignore nil error.
//
// Accepts errorable function result (v, err), returns v if err is nil,
// otherwise panics with err.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
