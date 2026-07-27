// Package holyapi provides client for interacting with other Holyheld services
// over HTTP
package holyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

// Request performs a simple JSON request to the endpoint (expects response to be JSON if
// successResponse arg is provided).
//
// Deprecated: Use [Client.RequestWithOptions] instead for better flexibility.
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

// RequestWithRetry allows [Client.Request] to have explicit retry policy.
//
// Deprecated: Use [Client.RequestWithOptions] instead for better readability.
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

// RequestWithOptions performs a JSON request to the endpoint (expects response to be JSON if
// successResponse arg is provided).
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

// FormDataRequest performs a multipart/formdata request to the endpoint
// (expects response to be JSON if successResponse arg is provided).
func (c *Client) FormDataRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	successResponse any,
	contentType string,
	opts ...rest.RequestOption,
) (int, error) {
	allOpts := make([]rest.RequestOption, 0, len(opts)+2)
	allOpts = append(allOpts,
		rest.WithBody(body),
		rest.WithContentType(contentType),
	)
	allOpts = append(allOpts, opts...)

	return c.requestInternal(
		ctx,
		method,
		path,
		nil,
		successResponse,
		allOpts...,
	)
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

	fullPath, err := buildURL(c.baseURL.Get(), path)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to parse url (base=%s, path=%s): %w",
			c.baseURL.Get(),
			path,
			err,
		)
	}

	allOpts := make([]rest.RequestOption, 0, len(opts)+2)
	allOpts = append(allOpts, rest.WithClient(c.httpClient))
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, rest.WithAdditionalHeaders(prepareHeader(ctx)))

	status, err := rest.Request(
		ctx,
		method,
		fullPath,
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
				URL:    fullPath,
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

func buildURL(baseURL, pathURL string) (string, error) {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL (raw=%s): %w", baseURL, err)
	}

	pathURL = strings.TrimPrefix(pathURL, "/")

	finalURL, err := base.Parse(pathURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative path (raw=%s): %w", pathURL, err)
	}

	return finalURL.String(), nil
}
