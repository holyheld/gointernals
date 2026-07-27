package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/holyheld/gointernals/pool"
)

type requestOptions struct {
	Client       *http.Client
	Retries      int
	CheckRetry   retryablehttp.CheckRetry
	Headers      http.Header
	Encoder      Encoder
	Decoder      Decoder
	Body         io.Reader
	ContentType  string
	ResponseHook ResponseHook
	StatusFilter StatusFilter
}

// ResponseHook is executed after receiving
// [*http.Response].
//
// Caller must not modify response body.
type ResponseHook func(r *http.Response)

// StatusFilter defines the range of response status codes that are
// accepted as "successful response". If condition fails, response is
// considered "failed".
type StatusFilter func(status int) bool

type RequestOption func(*requestOptions)

func WithClient(c *http.Client) RequestOption {
	return func(o *requestOptions) { o.Client = c }
}

func WithRetries(r int) RequestOption {
	return func(o *requestOptions) { o.Retries = r }
}

func WithCheckRetry(cr retryablehttp.CheckRetry) RequestOption {
	return func(o *requestOptions) { o.CheckRetry = cr }
}

func WithHeaders(h http.Header) RequestOption {
	return func(o *requestOptions) { o.Headers = h }
}

func WithAdditionalHeaders(h http.Header) RequestOption {
	return func(o *requestOptions) {
		if o.Headers == nil {
			o.Headers = make(http.Header)
		}

		CopyHeader(&o.Headers, h)
	}
}

func WithSerializer(s Serializer) RequestOption {
	return func(o *requestOptions) {
		o.Encoder = s
		o.Decoder = s
	}
}

func WithEncoder(e Encoder) RequestOption {
	return func(o *requestOptions) {
		o.Encoder = e
	}
}

func WithDecoder(d Decoder) RequestOption {
	return func(o *requestOptions) {
		o.Decoder = d
	}
}

func WithBody(r io.Reader) RequestOption {
	return func(o *requestOptions) {
		o.Body = r
	}
}

func WithContentType(ct string) RequestOption {
	return func(o *requestOptions) {
		o.ContentType = ct
	}
}

// WithResponseHook injects hook to be executed after receiving
// [*http.Response].
//
// Caller must not modify response body.
func WithResponseHook(h ResponseHook) RequestOption {
	return func(o *requestOptions) {
		o.ResponseHook = h
	}
}

func WithStatusFilter(f StatusFilter) RequestOption {
	return func(o *requestOptions) {
		o.StatusFilter = f
	}
}

func buildClient(opts *requestOptions) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	if opts.CheckRetry != nil {
		retryClient.CheckRetry = opts.CheckRetry
	}

	// disable internal logger (it leaks URLs to stdout by default)
	retryClient.Logger = nil
	retryClient.RetryMax = max(opts.Retries, 0)
	retryClient.HTTPClient = opts.Client
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler

	return retryClient
}

// JSONRequest performs JSON request to the endpoint.
//
// Deprecated: identical to [Request].
func JSONRequest(
	ctx context.Context,
	method string,
	url string,
	input any,
	output any,
	errorResp any,
	opts ...RequestOption,
) (int, error) {
	return request(
		ctx,
		method,
		url,
		input,
		output,
		errorResp,
		opts...,
	)
}

// Request performs a HTTP request to the endpoint.
//
// Request is made with default non-shared non-keepalive [*http.Client], if not specified otherwise
// by request options.
//
// If body is provided, assumes it to be Content-Type: application/json (if not specified otherwise
// by request options). [WithBody] takes precedence over input param, so the latter is ignored.
// Otherwise, with both input param is nil and [WithBody] not used, adds no Content-Type header
//
// If output (or errorResp) is provided, decodes response as Content-Type: application/json (if not
// specified otherwise by request options).
// Encodes input and decodes output with [DefaultSerializer], if not specified otherwise by
// request options.
//
// Writes to input if [StatusFilter] if satisfied, otherwise writes to errorResp.
// Does not read the response body is matching target is nil.
//
// Returns status code and a nil error if request was made successfully, otherwise
// returns 0 and the request error.
func Request(
	ctx context.Context,
	method string,
	url string,
	input any,
	output any,
	errorResp any,
	opts ...RequestOption,
) (int, error) {
	return request(
		ctx,
		method,
		url,
		input,
		output,
		errorResp,
		opts...,
	)
}

func request(
	ctx context.Context,
	method string,
	url string,
	input any,
	output any,
	errorResp any,
	opts ...RequestOption,
) (int, error) {
	options := &requestOptions{
		Retries:      0,
		CheckRetry:   nil,
		Headers:      make(http.Header),
		Encoder:      defaultSerializer,
		Decoder:      defaultSerializer,
		Client:       cleanhttp.DefaultClient(),
		Body:         nil,
		ContentType:  "",
		ResponseHook: nil,
		StatusFilter: defaultStatusFilter,
	}

	for _, opt := range opts {
		opt(options)
	}

	var body io.Reader = http.NoBody
	if options.Body != nil {
		body = options.Body
	} else if input != nil {
		pool := getPool(pool.Unsized)

		buf := pool.Get()
		defer pool.Put(buf)

		err := EncodeCustom(options.Encoder, buf, input)
		if err != nil {
			return 0, fmt.Errorf("failed to encode data: %w", err)
		}

		body = buf
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, NewRequestCreationError(method, url, body, err)
	}

	if options.ContentType != "" {
		req.Header.Set("Content-Type", options.ContentType)
	} else if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	cli := buildClient(options)

	CopyHeader(&req.Header, options.Headers)

	r, err := cli.Do(req)
	if err != nil {
		return 0, NewRequestExecutionError(method, url, body, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()

	target := output
	if !options.StatusFilter(r.StatusCode) {
		target = errorResp
	}

	if target == nil {
		return r.StatusCode, nil
	}

	contentLength := r.ContentLength
	// handle transfer-encoding
	if contentLength <= 0 {
		contentLength = defaultTransferEncodingBufferSize
	}

	pool := getPool(contentLength)

	buf := pool.Get()
	defer pool.Put(buf)

	tee := io.TeeReader(r.Body, buf)

	err = DecodeCustom(options.Decoder, tee, target)
	if err != nil {
		return r.StatusCode, NewRequestParsingError(
			method,
			url,
			body,
			buf,
			err,
		)
	}

	return r.StatusCode, nil
}

func defaultStatusFilter(status int) bool {
	return status >= 200 && status <= 299
}
