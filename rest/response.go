package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
)

type ResponseStatus string

const (
	APIResponseStatusOk    ResponseStatus = "ok"
	APIResponseStatusError ResponseStatus = "error"
)

type ErrorCode string

const (
	InternalServerErrorCode ErrorCode = "INTERNAL_SERVER_ERROR"
	BadRequestErrorCode     ErrorCode = "BAD_REQUEST"
	NotFoundErrorCode       ErrorCode = "NOT_FOUND"
	TimeoutErrorCode        ErrorCode = "TIMEOUT"
	UnauthorizedErrorCode   ErrorCode = "UNAUTHORIZED"
	ForbiddenErrorCode      ErrorCode = "FORBIDDEN"
	InvalidSessionErrorCode ErrorCode = "INVALID_SESSION"
	ConflictErrorCode       ErrorCode = "CONFLICT"
)

type Response struct {
	Status    ResponseStatus `json:"status"`
	ErrorCode ErrorCode      `json:"errorCode,omitempty"`
	Error     string         `json:"error,omitempty"`
	Payload   any            `json:"payload,omitempty"`
}

func JSONResponse(w http.ResponseWriter, resp *Response, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = Encode(w, resp)
}

func JSONResponseRaw(w http.ResponseWriter, payload any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = Encode(w, payload)
}

type RawResponse struct {
	Status int
	Body   *bytes.Buffer
	Header http.Header

	cleanup func()
}

func (r *RawResponse) Close() {
	if r.cleanup != nil {
		r.cleanup()
	}
}

func getETagFromBytes[T ~[]byte](b T) (string, error) {
	h := fnv.New128()

	_, err := h.Write(b)
	if err != nil {
		return "", fmt.Errorf("failed to write bytes to hash: %w", err)
	}

	return fmt.Sprintf("W/\"%x\"", h.Sum(nil)), nil
}

// ServeOKResponseJSON serves [http.StatusOK] [Response] with provided payload,
// optionally handling ETag match
//
// Will serve [http.StatusNotModified] with no body if provided ETag is equal to computed,
// otherwise will serve [http.StatusOK] with body.
func ServeOKResponseJSON(w http.ResponseWriter, r *http.Request, res any) {
	b, err := Marshal(res)
	if err != nil {
		JSONResponse(w, &Response{
			Status:  APIResponseStatusOk,
			Payload: res,
		}, http.StatusOK)

		return
	}

	etag, err := getETagFromBytes(b)
	if err != nil {
		JSONResponse(w, &Response{
			Status:  APIResponseStatusOk,
			Payload: json.RawMessage(b),
		}, http.StatusOK)

		return
	}

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)

		return
	}

	w.Header().Set("ETag", etag)

	JSONResponse(w, &Response{
		Status:  APIResponseStatusOk,
		Payload: json.RawMessage(b),
	}, http.StatusOK)
}

func ServeBadRequestMalformedPayload(w http.ResponseWriter) {
	ServeBadRequest(w, "Malformed payload")
}

func ServeBadRequest(w http.ResponseWriter, err string) {
	JSONResponse(w, &Response{
		Status:    APIResponseStatusError,
		ErrorCode: BadRequestErrorCode,
		Error:     err,
	}, http.StatusBadRequest)
}

func ServeInternalServerError(w http.ResponseWriter, err string) {
	JSONResponse(w, &Response{
		Status:    APIResponseStatusError,
		ErrorCode: InternalServerErrorCode,
		Error:     err,
	}, http.StatusInternalServerError)
}

func ServeNotFound(w http.ResponseWriter, err string) {
	JSONResponse(w, &Response{
		Status:    APIResponseStatusError,
		ErrorCode: NotFoundErrorCode,
		Error:     err,
	}, http.StatusNotFound)
}
