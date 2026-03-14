package holyapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type Status string

const (
	StatusOK    Status = "ok"
	StatusError Status = "error"
)

type ResponseSuccess struct {
	Status  Status `json:"status"`
	Payload any    `json:"payload,omitempty"`
}

type ResponseError struct {
	Status           Status          `json:"status"`
	ErrorDescription string          `json:"error"`
	ErrorCode        string          `json:"errorCode"`
	Payload          json.RawMessage `json:"payload,omitempty"`
	Meta             ResponseMeta    `json:"-"`
}

type ResponseMeta struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Status int    `json:"status"`
}

func (r *ResponseError) Error() string {
	errorCode := strings.TrimSpace(r.ErrorCode)
	if errorCode == "" {
		errorCode = "NO_ERROR_CODE"
	}

	errorDescription := strings.TrimSpace(r.ErrorDescription)
	if errorDescription == "" {
		errorDescription = "NO_ERROR_DESCRIPTION"
	}

	if len(r.Payload) > 0 {
		return fmt.Sprintf(
			"request error (%s %s => %d): %s: %s (%v)",
			r.Meta.Method,
			r.Meta.URL,
			r.Meta.Status,
			errorCode,
			errorDescription,
			string(r.Payload),
		)
	}

	return fmt.Sprintf(
		"request error (%s %s => %d): %s: %s",
		r.Meta.Method,
		r.Meta.URL,
		r.Meta.Status,
		errorCode,
		errorDescription,
	)
}

func (r *ResponseError) LogValue() slog.Value {
	errorCode := strings.TrimSpace(r.ErrorCode)
	if errorCode == "" {
		errorCode = "NO_ERROR_CODE"
	}

	errorDescription := strings.TrimSpace(r.ErrorDescription)
	if errorDescription == "" {
		errorDescription = "NO_ERROR_DESCRIPTION"
	}

	parts := []slog.Attr{
		slog.String("message", "request error"),
		slog.String("errorCode", errorCode),
		slog.String("error", errorDescription),
		slog.GroupAttrs(
			"requestMeta",
			slog.String("method", r.Meta.Method),
			slog.String("url", r.Meta.URL),
			slog.Int("status", r.Meta.Status),
		),
	}

	if r.Payload != nil {
		parts = append(parts, slog.String("payload", string(r.Payload)))
	}

	return slog.GroupValue(parts...)
}

type UnexpectedError struct {
	status int
	cause  *ResponseError
}

func NewUnexpectedError(status int, cause *ResponseError) *UnexpectedError {
	return &UnexpectedError{
		status: status,
		cause:  cause,
	}
}

func (e *UnexpectedError) Error() string {
	return fmt.Sprintf("unexpected error (status code %d): %s", e.status, e.cause.Error())
}

func (e *UnexpectedError) Is(target error) bool {
	return errors.Is(e.cause, target)
}

func (e *UnexpectedError) Unwrap() error {
	return e.cause
}

func (e *UnexpectedError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", "Unexpected error"),
		slog.Int("status", e.status),
		slog.Any("cause", e.cause.LogValue()),
	)
}

type NonContractResponseError struct {
	status int
	cause  error
}

func NewNonContractResponseError(status int, cause error) *NonContractResponseError {
	return &NonContractResponseError{
		status: status,
		cause:  cause,
	}
}

func (e *NonContractResponseError) Error() string {
	return fmt.Sprintf("unexpected error (status code %d): %s", e.status, e.cause.Error())
}

func (e *NonContractResponseError) Is(target error) bool {
	return errors.Is(e.cause, target)
}

func (e *NonContractResponseError) Unwrap() error {
	return e.cause
}

func (e *NonContractResponseError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", "Non-contract error"),
		slog.Int("status", e.status),
		slog.Any("cause", e.cause),
	)
}

type UnexpectedStatusError struct {
	status int
}

func NewUnexpectedStatusError(status int) *UnexpectedStatusError {
	return &UnexpectedStatusError{
		status: status,
	}
}

func (e *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("unexpected response status %d", e.status)
}

func (e *UnexpectedStatusError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", "Unexpected response status"),
		slog.Int("status", e.status),
	)
}
