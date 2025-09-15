package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()

	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.RetryCount)
	assert.Equal(t, 1*time.Second, config.RetryWaitTime)
	assert.Equal(t, 30*time.Second, config.RetryMaxWaitTime)
	assert.NotNil(t, config.RateLimiter)
	assert.Equal(t, "git-pr-automation/1.0", config.UserAgent)
	assert.False(t, config.Debug)
	assert.NotNil(t, config.Headers)
	assert.Equal(t, 0, len(config.Headers))
}

func TestNewHTTPClient_WithConfig(t *testing.T) {
	config := &HTTPClientConfig{
		BaseURL:          "https://api.example.com",
		Timeout:          15 * time.Second,
		RetryCount:       5,
		RetryWaitTime:    2 * time.Second,
		RetryMaxWaitTime: 60 * time.Second,
		RateLimiter:      rate.NewLimiter(rate.Every(100*time.Millisecond), 5),
		UserAgent:        "custom-agent/2.0",
		Debug:            true,
		Headers: map[string]string{
			"Custom-Header": "custom-value",
			"X-API-Key":     "secret-key",
		},
	}

	client := NewHTTPClient(config)

	assert.NotNil(t, client)
	// Note: Due to resty's internal structure, we can't directly test all config values
	// but we can verify the client was created without panic
}

func TestNewHTTPClient_NilConfig(t *testing.T) {
	client := NewHTTPClient(nil)

	assert.NotNil(t, client)
	// Should use default configuration
}

func TestRateLimitedHTTPClient(t *testing.T) {
	requestsPerSecond := 10.0
	burst := 20

	client := RateLimitedHTTPClient(requestsPerSecond, burst)

	assert.NotNil(t, client)
}

func TestNewHTTPClientFromConfig(t *testing.T) {
	config := HTTPClientConfig{
		BaseURL:   "https://api.test.com",
		Timeout:   10 * time.Second,
		UserAgent: "test-agent/1.0",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	httpClient := NewHTTPClientFromConfig(config)

	assert.NotNil(t, httpClient)
	assert.NotNil(t, httpClient.client)
}

func TestHTTPClient_SetBasicAuth(t *testing.T) {
	config := HTTPClientConfig{}
	httpClient := NewHTTPClientFromConfig(config)

	// Should not panic
	assert.NotPanics(t, func() {
		httpClient.SetBasicAuth("username", "password")
	})
}

func TestHTTPClient_SetAuthToken(t *testing.T) {
	config := HTTPClientConfig{}
	httpClient := NewHTTPClientFromConfig(config)

	// Should not panic
	assert.NotPanics(t, func() {
		httpClient.SetAuthToken("bearer-token")
	})
}

func TestHTTPClient_SetHeader(t *testing.T) {
	config := HTTPClientConfig{}
	httpClient := NewHTTPClientFromConfig(config)

	// Should not panic
	assert.NotPanics(t, func() {
		httpClient.SetHeader("X-Custom-Header", "custom-value")
	})
}

func TestHTTPClient_Get_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/test-endpoint", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "success", "data": "test-data"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Get(ctx, "/test-endpoint", &result)

	assert.NoError(t, err)
	assert.Equal(t, "success", result["message"])
	assert.Equal(t, "test-data", result["data"])
}

func TestHTTPClient_Get_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "bad request"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Get(ctx, "/error-endpoint", &result)

	assert.Error(t, err)

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	assert.Contains(t, httpErr.Message, "bad request")
}

func TestHTTPClient_Post_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/create", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id": 123, "status": "created"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	requestBody := map[string]interface{}{
		"name":        "test-item",
		"description": "test description",
	}
	var result map[string]interface{}

	err := httpClient.Post(ctx, "/create", requestBody, &result)

	assert.NoError(t, err)
	assert.Equal(t, float64(123), result["id"]) // JSON numbers are parsed as float64
	assert.Equal(t, "created", result["status"])
}

func TestHTTPClient_Post_NoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	err := httpClient.Post(ctx, "/no-body", nil, nil)

	assert.NoError(t, err)
}

func TestHTTPClient_Post_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"error": "validation failed"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	requestBody := map[string]string{"invalid": "data"}

	err := httpClient.Post(ctx, "/validation-error", requestBody, nil)

	assert.Error(t, err)

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.StatusCode)
	assert.Contains(t, httpErr.Message, "validation failed")
}

func TestHTTPClient_Put_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/update/123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"id": 123, "status": "updated"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	requestBody := map[string]string{"name": "updated-name"}
	var result map[string]interface{}

	err := httpClient.Put(ctx, "/update/123", requestBody, &result)

	assert.NoError(t, err)
	assert.Equal(t, float64(123), result["id"])
	assert.Equal(t, "updated", result["status"])
}

func TestHTTPClient_Put_NoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	err := httpClient.Put(ctx, "/update", nil, nil)

	assert.NoError(t, err)
}

func TestHTTPClient_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/delete/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	err := httpClient.Delete(ctx, "/delete/123", nil)

	assert.NoError(t, err)
}

func TestHTTPClient_Delete_WithResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"deleted": true, "id": 123}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Delete(ctx, "/delete/123", &result)

	assert.NoError(t, err)
	assert.Equal(t, true, result["deleted"])
	assert.Equal(t, float64(123), result["id"])
}

func TestHTTPClient_Delete_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "resource not found"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	err := httpClient.Delete(ctx, "/delete/999", nil)

	assert.Error(t, err)

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusNotFound, httpErr.StatusCode)
	assert.Contains(t, httpErr.Message, "resource not found")
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Longer than context timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second, // Longer than context timeout
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := httpClient.Get(ctx, "/slow-endpoint", nil)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, duration, 50*time.Millisecond) // Should timeout quickly
}

func TestHTTPClient_ServerTimeout(t *testing.T) {
	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Longer than client timeout
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 50 * time.Millisecond, // Very short timeout
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	start := time.Now()
	err := httpClient.Get(ctx, "/timeout-endpoint", nil)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, duration, 40*time.Millisecond) // Should respect timeout
	assert.Less(t, duration, 100*time.Millisecond)          // But not wait too long
}

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{
		StatusCode: http.StatusBadRequest,
		Message:    "Invalid request format",
	}

	assert.Equal(t, "Invalid request format", err.Error())
}

func TestHTTPError_GetStatusCode(t *testing.T) {
	err := &HTTPError{
		StatusCode: http.StatusNotFound,
		Message:    "Resource not found",
	}

	assert.Equal(t, http.StatusNotFound, err.GetStatusCode())
}

func TestHTTPClient_RetryConditions(t *testing.T) {
	// Test that the client retries on appropriate HTTP status codes
	retryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 3 {
			// Return a retryable error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"error": "service temporarily unavailable"}`)
		} else {
			// Finally succeed
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"success": true}`)
		}
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL:          server.URL,
		RetryCount:       3,
		RetryWaitTime:    10 * time.Millisecond,
		RetryMaxWaitTime: 50 * time.Millisecond,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Get(ctx, "/retry-test", &result)

	assert.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, 3, retryCount) // Should have retried twice, then succeeded on third attempt
}

func TestHTTPClient_NoRetryOnClientError(t *testing.T) {
	// Test that the client does not retry on 4xx client errors (except 429)
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest) // 400 - should not retry
		fmt.Fprintf(w, `{"error": "bad request"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL:       server.URL,
		RetryCount:    3,
		RetryWaitTime: 10 * time.Millisecond,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	err := httpClient.Get(ctx, "/client-error", nil)

	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Should not retry 400 errors

	var httpErr *HTTPError
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
}

func TestHTTPClient_RetryOnTooManyRequests(t *testing.T) {
	// Test that the client retries on 429 Too Many Requests
	retryCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests) // 429 - should retry
			fmt.Fprintf(w, `{"error": "rate limited"}`)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"success": true}`)
		}
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL:       server.URL,
		RetryCount:    3,
		RetryWaitTime: 10 * time.Millisecond,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Get(ctx, "/rate-limit-test", &result)

	assert.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, 2, retryCount) // Should retry once on 429, then succeed
}

func TestHTTPClient_WithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers are present
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.api+json", r.Header.Get("Accept"))
		assert.Equal(t, "custom-agent/3.0", r.Header.Get("User-Agent"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"received": "with headers"}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{
		BaseURL:   server.URL,
		UserAgent: "custom-agent/3.0",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"Accept":        "application/vnd.api+json",
		},
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()
	var result map[string]interface{}

	err := httpClient.Get(ctx, "/headers-test", &result)

	assert.NoError(t, err)
	assert.Equal(t, "with headers", result["received"])
}

func TestHTTPClient_RateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	}))
	defer server.Close()

	// Very restrictive rate limiting for testing
	rateLimiter := rate.NewLimiter(rate.Every(50*time.Millisecond), 1)

	config := HTTPClientConfig{
		BaseURL:     server.URL,
		RateLimiter: rateLimiter,
	}
	httpClient := NewHTTPClientFromConfig(config)

	ctx := context.Background()

	// First request should be immediate
	start := time.Now()
	err1 := httpClient.Get(ctx, "/rate-limited", nil)
	firstDuration := time.Since(start)

	assert.NoError(t, err1)
	assert.Less(t, firstDuration, 20*time.Millisecond)

	// Second request should be rate limited
	start = time.Now()
	err2 := httpClient.Get(ctx, "/rate-limited", nil)
	secondDuration := time.Since(start)

	assert.NoError(t, err2)
	assert.GreaterOrEqual(t, secondDuration, 40*time.Millisecond) // Should be rate limited
}

// Benchmark tests
func BenchmarkHTTPClient_Get(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"benchmark": true}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		err := httpClient.Get(ctx, "/benchmark", &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPClient_Post(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"created": true}`)
	}))
	defer server.Close()

	config := HTTPClientConfig{BaseURL: server.URL}
	httpClient := NewHTTPClientFromConfig(config)
	ctx := context.Background()

	requestBody := map[string]string{"test": "data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		err := httpClient.Post(ctx, "/benchmark", requestBody, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
