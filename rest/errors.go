package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

type phase = string

const (
	creation        phase = "creation"
	execution       phase = "execution"
	responseParsing phase = "response parsing"
)

type RequestFailedError struct {
	phase           phase
	method          string
	url             string
	requestPayload  string
	responsePayload string
	status          int
	err             error
}

func NewRequestCreationError(
	method string,
	url string,
	payload io.Reader,
	err error,
) *RequestFailedError {
	pb, _ := io.ReadAll(payload)

	return &RequestFailedError{
		method:         method,
		phase:          creation,
		url:            url,
		requestPayload: string(pb),
		err:            err,
	}
}

func NewRequestExecutionError(
	method string,
	url string,
	payload io.Reader,
	err error,
) *RequestFailedError {
	pb, _ := io.ReadAll(payload)

	return &RequestFailedError{
		method:         method,
		phase:          execution,
		url:            url,
		requestPayload: string(pb),
		err:            err,
	}
}

func NewRequestParsingError(
	method string,
	url string,
	requestPayload io.Reader,
	responsePayload io.Reader,
	err error,
) *RequestFailedError {
	reqpb, _ := io.ReadAll(requestPayload)
	respb, _ := io.ReadAll(responsePayload)

	return &RequestFailedError{
		method:          method,
		phase:           responseParsing,
		url:             url,
		requestPayload:  string(reqpb),
		responsePayload: string(respb),
		err:             err,
	}
}

func (e *RequestFailedError) Error() string {
	if e.requestPayload == "" {
		return fmt.Sprintf(
			"request to %s failed during %s: %s",
			e.url,
			e.phase,
			e.err,
		)
	}

	return fmt.Sprintf(
		"request to %s with data %v failed during %s: %s",
		e.url,
		e.requestPayload,
		e.phase,
		e.err,
	)
}

func (e *RequestFailedError) Is(target error) bool {
	return errors.Is(e.err, target)
}

func (e *RequestFailedError) Unwrap() error {
	return e.err
}

func (e *RequestFailedError) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("message", "Request failed during "+e.phase),
		slog.String("method", e.method),
		slog.String("url", e.url),
		slog.Any("cause", e.err),
	}

	if e.status != 0 {
		attrs = append(attrs, slog.Int("status", e.status))
	}

	if e.requestPayload != "" {
		attrs = append(attrs, slog.String("requestPayload", e.requestPayload))
	}

	if e.responsePayload != "" {
		attrs = append(attrs, slog.String("responsePayload", e.responsePayload))
	}

	return slog.GroupValue(attrs...)
}

type ValidationError struct {
	Field  string          `json:"field"`
	Reason string          `json:"reason"`
	Meta   json.RawMessage `json:"meta,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Reason)
}

func (e *ValidationError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", "Validation failed"),
		slog.String("field", e.Field),
		slog.String("reason", e.Reason),
	)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	messages := strings.Builder{}

	fmt.Fprintf(&messages, "%d validation errors occured\n", len(e))

	for _, err := range e {
		fmt.Fprintf(&messages, "%s\n", err)
	}

	return messages.String()
}

func (e ValidationErrors) Unwrap() []error {
	res := make([]error, len(e))

	for i, err := range e {
		res[i] = &err
	}

	return res
}

func (e ValidationErrors) LogValue() slog.Value {
	attrs := make([]slog.Attr, len(e))

	for idx, err := range e {
		attrs[idx] = slog.Any(strconv.FormatInt(int64(idx), 10), err)
	}

	return slog.GroupValue(attrs...)
}
