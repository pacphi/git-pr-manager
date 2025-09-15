package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

// HTTPClientConfig contains configuration for HTTP clients
type HTTPClientConfig struct {
	BaseURL          string
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	RateLimiter      *rate.Limiter
	UserAgent        string
	Debug            bool
	Headers          map[string]string
}

// DefaultHTTPClientConfig returns a default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 30 * time.Second,
		RateLimiter:      rate.NewLimiter(rate.Every(200*time.Millisecond), 10), // 5 req/sec, burst 10
		UserAgent:        "git-pr-automation/1.0",
		Debug:            false,
		Headers:          make(map[string]string),
	}
}

// NewHTTPClient creates a new HTTP client with the specified configuration
func NewHTTPClient(config *HTTPClientConfig) *resty.Client {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	client := resty.New()

	// Basic configuration
	if config.BaseURL != "" {
		client.SetBaseURL(config.BaseURL)
	}
	client.SetTimeout(config.Timeout)
	client.SetRetryCount(config.RetryCount)
	client.SetRetryWaitTime(config.RetryWaitTime)
	client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)
	client.SetHeader("User-Agent", config.UserAgent)
	client.SetDebug(config.Debug)

	// Set custom headers
	for key, value := range config.Headers {
		client.SetHeader(key, value)
	}

	// Add rate limiting if configured
	if config.RateLimiter != nil {
		client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
			if err := config.RateLimiter.Wait(req.Context()); err != nil {
				return fmt.Errorf("rate limiter error: %w", err)
			}
			return nil
		})
	}

	// Add request/response logging
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		logger := GetGlobalLogger().WithFields(map[string]interface{}{
			"method": req.Method,
			"url":    req.URL,
		})
		logger.Debug("Making HTTP request")
		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		logger := GetGlobalLogger().WithFields(map[string]interface{}{
			"method":      resp.Request.Method,
			"url":         resp.Request.URL,
			"status_code": resp.StatusCode(),
			"duration":    resp.Time(),
		})

		if resp.IsError() {
			logger.Warn("HTTP request failed")
		} else {
			logger.Debug("HTTP request completed")
		}
		return nil
	})

	// Configure retry conditions
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return true // Retry on network errors
		}

		// Retry on 5xx errors and specific 4xx errors
		switch r.StatusCode() {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		default:
			return false
		}
	})

	// Note: OnRetry is not available in this version of resty
	// Rate limiting is handled by the rate limiter middleware above

	return client
}

// RateLimitedHTTPClient creates an HTTP client with specific rate limiting
func RateLimitedHTTPClient(requestsPerSecond float64, burst int) *resty.Client {
	config := DefaultHTTPClientConfig()
	config.RateLimiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
	return NewHTTPClient(config)
}

// HTTPClient provides a simple interface for HTTP operations
type HTTPClient struct {
	client *resty.Client
}

// NewHTTPClientFromConfig creates a new HTTPClient with the specified configuration
func NewHTTPClientFromConfig(config HTTPClientConfig) *HTTPClient {
	configPtr := &config
	client := NewHTTPClient(configPtr)
	return &HTTPClient{
		client: client,
	}
}

// SetBasicAuth sets basic authentication for the HTTP client
func (h *HTTPClient) SetBasicAuth(username, password string) {
	h.client.SetBasicAuth(username, password)
}

// SetAuthToken sets bearer token authentication for the HTTP client
func (h *HTTPClient) SetAuthToken(token string) {
	h.client.SetAuthToken(token)
}

// SetHeader sets a header for all requests
func (h *HTTPClient) SetHeader(key, value string) {
	h.client.SetHeader(key, value)
}

// Get performs a GET request and unmarshals the response into the result
func (h *HTTPClient) Get(ctx context.Context, path string, result interface{}) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(result).
		Get(path)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    string(resp.Body()),
		}
	}

	return nil
}

// Post performs a POST request with the given body and unmarshals the response
func (h *HTTPClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	req := h.client.R().SetContext(ctx)

	if body != nil {
		req = req.SetBody(body)
	}

	if result != nil {
		req = req.SetResult(result)
	}

	resp, err := req.Post(path)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    string(resp.Body()),
		}
	}

	return nil
}

// Put performs a PUT request with the given body and unmarshals the response
func (h *HTTPClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	req := h.client.R().SetContext(ctx)

	if body != nil {
		req = req.SetBody(body)
	}

	if result != nil {
		req = req.SetResult(result)
	}

	resp, err := req.Put(path)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    string(resp.Body()),
		}
	}

	return nil
}

// Delete performs a DELETE request
func (h *HTTPClient) Delete(ctx context.Context, path string, result interface{}) error {
	req := h.client.R().SetContext(ctx)

	if result != nil {
		req = req.SetResult(result)
	}

	resp, err := req.Delete(path)
	if err != nil {
		return err
	}

	if resp.IsError() {
		return &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    string(resp.Body()),
		}
	}

	return nil
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return e.Message
}

// GetStatusCode returns the HTTP status code
func (e *HTTPError) GetStatusCode() int {
	return e.StatusCode
}
