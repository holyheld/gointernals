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
	Client     *http.Client
	Retries    int
	CheckRetry retryablehttp.CheckRetry
	Headers    http.Header
	Encoder    Encoder
	Decoder    Decoder
}

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

func buildClient(opts *requestOptions) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	if opts.CheckRetry != nil {
		retryClient.CheckRetry = opts.CheckRetry
	}

	// disable internal logger (it leaks URLs to stdout by default)
	retryClient.Logger = nil
	retryClient.RetryMax = max(opts.Retries, 0)
	retryClient.HTTPClient = opts.Client

	return retryClient
}

func JSONRequest(
	ctx context.Context,
	method string,
	url string,
	input any,
	output any,
	errorResp any,
	opts ...RequestOption,
) (int, error) {
	options := &requestOptions{
		Retries:    0,
		CheckRetry: nil,
		Headers:    make(http.Header),
		Encoder:    defaultSerializer,
		Decoder:    defaultSerializer,
		Client:     cleanhttp.DefaultClient(),
	}

	for _, opt := range opts {
		opt(options)
	}

	var body io.Reader = http.NoBody

	if input != nil {
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

	if input != nil {
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
	if r.StatusCode > 299 || r.StatusCode < 200 {
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
