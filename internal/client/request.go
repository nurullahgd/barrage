package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type ErrorCategory int

const (
	CategoryNone ErrorCategory = iota
	CategoryTimeout
	CategoryConnection
	CategoryClientError
	CategoryServerError
)

type Result struct {
	Error      error
	Latency    time.Duration
	StatusCode int
	Category   ErrorCategory
}

func DoRequest(httpClient *http.Client, url string, method string, body []byte) Result {
	start := time.Now()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return Result{Error: err, Latency: time.Since(start), Category: CategoryConnection}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		category := CategoryConnection
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			category = CategoryTimeout
		}
		return Result{Error: err, Latency: time.Since(start), Category: category}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()
	_, _ = io.Copy(io.Discard, resp.Body)

	category := CategoryNone
	switch {
	case resp.StatusCode >= 500:
		category = CategoryServerError
	case resp.StatusCode >= 400:
		category = CategoryClientError
	}

	return Result{Latency: time.Since(start), StatusCode: resp.StatusCode, Category: category}
}
