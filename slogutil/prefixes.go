package slogutil

import "log/slog"

const (
	moduleKey   = "module"
	functionKey = "function"
	payloadKey  = "payload"
	routineKey  = "routine"
)

func Module(name string) slog.Attr {
	return slog.String(moduleKey, name)
}

func Function(name string) slog.Attr {
	return slog.String(functionKey, name)
}

func Payload(args ...slog.Attr) slog.Attr {
	return slog.GroupAttrs(payloadKey, args...)
}

func Routine(name string) slog.Attr {
	return slog.String(routineKey, name)
}
