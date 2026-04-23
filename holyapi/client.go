// Package holyapi provides client for interacting with other Holyheld services
// over HTTP
package holyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/holyheld/gointernals/holder"
	"github.com/holyheld/gointernals/rest"
)

type Client struct {
	baseURL    holder.Holder[string]
	httpClient *http.Client
}

type Option func(*Client)

func WithClient(c *http.Client) Option {
	return func(o *Client) {
		o.httpClient = c
	}
}

func NewClient(baseURL holder.Holder[string], opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		httpClient: cleanhttp.DefaultPooledClient(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Request(
	ctx context.Context,
	method string,
	path string,
	header http.Header,
	body any,
	successResponse any,
) (int, error) {
	return c.requestInternal(
		ctx,
		method,
		path,
		body,
		successResponse,
		rest.WithHeaders(header),
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
	checkRetry retryablehttp.CheckRetry,
) (int, error) {
	return c.requestInternal(
		ctx,
		method,
		path,
		body,
		successResponse,
		rest.WithHeaders(header),
		rest.WithRetries(retries),
		rest.WithCheckRetry(checkRetry),
	)
}

func (c *Client) RequestWithOptions(
	ctx context.Context,
	method string,
	path string,
	body any,
	successResponse any,
	opts ...rest.RequestOption,
) (int, error) {
	return c.requestInternal(ctx, method, path, body, successResponse, opts...)
}

func (c *Client) requestInternal(
	ctx context.Context,
	method string,
	path string,
	body any,
	successResponse any,
	opts ...rest.RequestOption,
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

	allOpts := make([]rest.RequestOption, 0, len(opts)+2)
	allOpts = append(allOpts, rest.WithClient(c.httpClient))
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, rest.WithAdditionalHeaders(prepareHeader(ctx)))

	status, err := rest.JSONRequest(
		ctx,
		method,
		pathURL.String(),
		body,
		successResp,
		errorResp,
		allOpts...,
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

func prepareHeader(ctx context.Context) http.Header {
	out := make(http.Header, 2)

	if via := ExtractVia(ctx); via != "" {
		out.Set("Via", via)
	}

	if ua := ExtractUserAgent(ctx); ua != "" {
		out.Set("User-Agent", ua)
	}

	return out
}
