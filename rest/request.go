package rest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/holyheld/pool"
)

var sharedTransport = &http.Transport{
	MaxIdleConns:        1000,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
}

func getClient(retries int, checkRetry retryablehttp.CheckRetry) *http.Client {
	if retries <= 0 {
		return &http.Client{
			Transport: sharedTransport,
			Timeout:   90 * time.Second,
		}
	}

	retryClient := retryablehttp.NewClient()
	if checkRetry == nil {
		retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy
	} else {
		retryClient.CheckRetry = checkRetry
	}

	// disable internal logger (it leaks URLs to stdout by default)
	retryClient.Logger = nil
	retryClient.RetryMax = retries
	retryClient.HTTPClient.Transport = sharedTransport

	return retryClient.StandardClient()
}

func JSONRequestAdvancedCustom(
	ctx context.Context,
	method string,
	url string,
	headers http.Header,
	input any,
	output any,
	errorResp any,
	retries int,
	checkRetry retryablehttp.CheckRetry,
) (int, error) {
	var body io.Reader = http.NoBody

	if input != nil {
		pool := getPool(pool.Unsized)

		buf := pool.Get()
		defer pool.Put(buf)

		err := Encode(buf, input)
		if err != nil {
			return 0, fmt.Errorf("failed to encode data: %w", err)
		}

		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, NewRequestCreationError(method, url, body, err)
	}

	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	cli := getClient(retries, checkRetry)

	CopyHeader(&req.Header, headers)

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

	err = Decode(tee, target)
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
