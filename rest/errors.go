package rest

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
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

type ValidationErrors map[string][]string

func NewValidationErrors() ValidationErrors {
	return make(ValidationErrors)
}

func (v ValidationErrors) Add(field, message string) {
	v[field] = append(v[field], message)
}

func (v ValidationErrors) CopyFrom(errs ValidationErrors) {
	for field, group := range errs {
		v[field] = append(v[field], group...)
	}
}

func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

func (v ValidationErrors) Error() string {
	var summary []string

	for field, msgs := range v {
		summary = append(summary, fmt.Sprintf("%s: [%s]", field, strings.Join(msgs, ", ")))
	}

	return fmt.Sprintf("validation failed: %s", strings.Join(summary, "; "))
}

func (v ValidationErrors) MarshalJSON() ([]byte, error) {
	return Marshal(map[string][]string(v))
}

func (v ValidationErrors) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(v))

	for field, msgs := range v {
		anyMsgs := make([]any, len(msgs))

		for i, m := range msgs {
			anyMsgs[i] = m
		}

		attrs = append(attrs, slog.Any(field, anyMsgs))
	}

	return slog.GroupValue(attrs...)
}
