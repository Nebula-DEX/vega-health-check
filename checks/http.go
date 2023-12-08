package checks

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	ErrHTTPFailedRequest error = errors.New("failed to send http request")
	ErrHTTPFailReadBody  error = errors.New("failed to read http response body")
	ErrHTTPFailUnmarshal error = errors.New("failed to unmarshal http response body")
)

type HTTPChecker struct {
	client *http.Client
	url    string
}

type HTTPCheckResult struct {
	StatusCode int
	Duration   time.Duration
	Headers    http.Header
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func NewHTTPChecker(endpoint string, httpClient *http.Client) *HTTPChecker {
	checker := &HTTPChecker{
		url:    endpoint,
		client: httpClient,
	}

	if httpClient == nil {
		checker.client = defaultHTTPClient()
	}

	return checker
}

func (c *HTTPChecker) Get(jsonOutput interface{}) (*HTTPCheckResult, error) {
	start := time.Now()
	resp, err := c.client.Get(c.url)
	duration := time.Since(start)

	if err != nil {
		return nil, errors.Join(ErrHTTPFailedRequest, err)
	}

	result := &HTTPCheckResult{
		Duration:   duration,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, errors.Join(ErrHTTPFailReadBody, err)
	}

	if jsonOutput != nil {
		if err := json.Unmarshal(b, jsonOutput); err != nil {
			return result, errors.Join(ErrHTTPFailUnmarshal, err)
		}
	}

	return result, nil
}
