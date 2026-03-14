package slogutil

import (
	"log/slog"
	"strconv"
)

const (
	errorKey   = "error"
	causeKey   = "cause"
	messageKey = "message"
)

// Error wraps provided error to named [slog.Attr], applies formatting for
// better error chain visibility.
func Error(err error) slog.Attr {
	return logError(errorKey, err)
}

func logError(key string, err error) slog.Attr {
	if err == nil {
		return slog.Any(key, err)
	}

	// if provides self LogValue() method, use it
	_, ok := err.(slog.LogValuer)
	if ok {
		return slog.Any(key, err)
	}

	attrs := []slog.Attr{slog.String(messageKey, err.Error())}

	// if provides Unwrap() error method, use it as cause (if present)
	wrapper, ok := err.(interface {
		Unwrap() error
	})
	if ok {
		wrapped := wrapper.Unwrap()
		if wrapped == nil {
			return slog.Any(key, err)
		}

		attrs = append(attrs, logError(causeKey, wrapped))
	}

	// if provides Unwrap() []error method, use it as slice of causes (if present)
	sliceWrapper, ok := err.(interface {
		Unwrap() []error
	})
	if ok {
		wrapped := sliceWrapper.Unwrap()
		if len(wrapped) == 0 {
			return slog.Any(key, err)
		}

		group := make([]slog.Attr, len(wrapped))
		for idx, err := range wrapped {
			group[idx] = logError(strconv.FormatInt(int64(idx), 10), err)
		}

		attrs = append(attrs, slog.GroupAttrs(causeKey, group...))
	}

	return slog.GroupAttrs(
		key,
		attrs...,
	)
}

const recoverKey = "recover"

// Recover wraps recover() return (typically an error, but sometimes used as any, as panic can
// use arbitrary data).
func Recover(r any) slog.Attr {
	switch r := r.(type) {
	case error:
		return logError(recoverKey, r)
	case string:
		return slog.String(recoverKey, r)
	}

	return slog.Any(recoverKey, r)
}
