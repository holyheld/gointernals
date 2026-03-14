package slogutil

import (
	"fmt"
	"log/slog"
	"reflect"
)

// StringLike tries with best efforts to represent the type as a [slog.AnyValue]
// attribute, and falls back to [slog.AnyValue] if unable to do so.
func StringLike(key string, value any) slog.Attr {
	switch v := (value).(type) {
	case string:
		return slog.String(key, v)
	case fmt.Stringer:
		return slog.String(key, v.String())
	case fmt.GoStringer:
		return slog.String(key, v.GoString())
	default:
		if reflect.TypeOf(v).ConvertibleTo(reflect.TypeFor[string]()) {
			return slog.String(key, reflect.ValueOf(v).String())
		}

		return slog.Any(key, value)
	}
}
