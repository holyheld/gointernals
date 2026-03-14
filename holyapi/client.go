// Package holyapi provides client for interacting with other Holyheld services
// over HTTP
package holyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/holyheld/holder"
	"github.com/holyheld/rest"
)

type Client struct {
	baseURL holder.Holder[string]
}

func NewClient(baseURL holder.Holder[string]) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

func (c *Client) Request(
	ctx context.Context,
	method string,
	path string,
	header http.Header,
	body any,
	successResponse any,
) (int, error) {
	return c.RequestWithRetry(
		ctx,
		method,
		path,
		header,
		body,
		successResponse,
		0,
		nil,
	)
}

func (c *Client) RequestWithRetry(
	ctx context.Context,
	method string,
	path string,
	header http.Header,
	body any,
	successResponse any,
	retries int,
	checkRetry func(ctx context.Context, resp *http.Response, err error) (bool, error),
) (int, error) {
	successResp := &ResponseSuccess{Payload: successResponse}
	if successResponse == nil {
		// if we don't want payload in the calling function, we still need to provide at least some type
		// due to assumption in the decoder that "any" type with value "nil" represents map[string]any,
		// which is significantly slower than copying bytes into dummy json.RawMessage ¯\(°_o)/¯
		successResp.Payload = &json.RawMessage{}
	}

	errorResp := &ResponseError{}

	baseURL, err := url.Parse(c.baseURL.Get())
	if err != nil {
		return 0, fmt.Errorf("failed to parse base url (raw=%s): %w", c.baseURL.Get(), err)
	}

	pathURL, err := baseURL.Parse(path)
	if err != nil {
		return 0, fmt.Errorf("failed to parse path url (raw=%s): %w", path, err)
	}

	status, err := rest.JSONRequestAdvancedCustom(
		ctx,
		method,
		pathURL.String(),
		prepareHeader(ctx, header),
		body,
		successResp,
		errorResp,
		retries,
		checkRetry,
	)
	if err != nil {
		return status, err
	}

	// Sometimes we make mistakes and regardless of HTTP status being 200 OK
	// respond with { "status": "error" }. If that happens, we will catch the error
	// properly below
	effectiveStatus := successResp.Status
	if status > 299 || status < 200 {
		effectiveStatus = errorResp.Status
	}

	if effectiveStatus == StatusError {
		return status, &ResponseError{
			Status:           StatusError,
			ErrorCode:        errorResp.ErrorCode,
			ErrorDescription: errorResp.ErrorDescription,
			Payload:          errorResp.Payload,
			Meta: ResponseMeta{
				URL:    pathURL.String(),
				Method: method,
				Status: status,
			},
		}
	}

	return status, nil
}

func prepareHeader(ctx context.Context, h http.Header) http.Header {
	out := h.Clone()
	if out == nil {
		out = make(http.Header)
	}

	if via := ExtractVia(ctx); via != "" {
		out.Add("Via", via)
	}

	if ua := ExtractUserAgent(ctx); ua != "" {
		out.Set("User-Agent", ua)
	}

	return out
}
