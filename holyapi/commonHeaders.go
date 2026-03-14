package holyapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type ctxKey int

const (
	ctxKeyVia ctxKey = iota
	ctxKeyUA
)

func ProvideVia(pseudonym string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			values := make([]string, 0)

			headerValue := r.Header.Get("Via")
			if headerValue != "" {
				values = append(values, r.Header.Get("Via"))
			}

			ctxValue := ExtractVia(ctx)
			if ctxValue != "" {
				values = append(values, ctxValue)
			}

			values = append(values, fmt.Sprintf("%s %s", r.Proto, pseudonym))

			ctx = context.WithValue(ctx, ctxKeyVia, strings.Join(values, ", "))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideUserAgent(ua string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(
				w,
				r.WithContext(context.WithValue(
					r.Context(),
					ctxKeyUA,
					ua,
				)),
			)
		})
	}
}

func ExtractVia(ctx context.Context) string {
	v, ok := ctx.Value(ctxKeyVia).(string)
	if ok {
		return v
	}

	return ""
}

func ExtractUserAgent(ctx context.Context) string {
	v, ok := ctx.Value(ctxKeyUA).(string)
	if ok {
		return v
	}

	return ""
}
