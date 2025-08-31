package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs request details", func(t *testing.T) {
		// Capture log output
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput) // Reset to original

		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond) // Small delay to test timing
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap with logging middleware
		wrappedHandler := loggingMiddleware(testHandler)

		// Create test request
		req := httptest.NewRequest("GET", "/test/path", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rr := httptest.NewRecorder()

		// Execute request
		wrappedHandler.ServeHTTP(rr, req)

		// Check response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		if rr.Body.String() != "test response" {
			t.Errorf("Expected 'test response', got '%s'", rr.Body.String())
		}

		// Check log output
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "GET") {
			t.Error("Expected log to contain HTTP method")
		}
		if !strings.Contains(logOutput, "/test/path") {
			t.Error("Expected log to contain request path")
		}
		if !strings.Contains(logOutput, "127.0.0.1:12345") {
			t.Error("Expected log to contain remote address")
		}
		// Check that timing is included (should contain time units)
		if !strings.Contains(logOutput, "ms") && !strings.Contains(logOutput, "µs") &&
			!strings.Contains(logOutput, "ns") && !strings.Contains(logOutput, "s") {
			t.Error("Expected log to contain timing information")
		}
	})

	t.Run("logs different HTTP methods", func(t *testing.T) {
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})

		wrappedHandler := loggingMiddleware(testHandler)

		// Test POST request
		req := httptest.NewRequest("POST", "/api/tasks", strings.NewReader(`{"title":"test"}`))
		req.RemoteAddr = "192.168.1.1:8080"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "POST") {
			t.Error("Expected log to contain POST method")
		}
		if !strings.Contains(logOutput, "/api/tasks") {
			t.Error("Expected log to contain request path")
		}
		if !strings.Contains(logOutput, "192.168.1.1:8080") {
			t.Error("Expected log to contain remote address")
		}
	})

	t.Run("logs query parameters", func(t *testing.T) {
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := loggingMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/api/tasks?status=done&limit=10", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "/api/tasks?status=done&limit=10") {
			t.Error("Expected log to contain full request URI with query parameters")
		}
	})

	t.Run("handles different status codes", func(t *testing.T) {
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput)

		testCases := []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		}

		for _, statusCode := range testCases {
			logBuffer.Reset()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			})

			wrappedHandler := loggingMiddleware(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, rr.Code)
			}

			// Verify request was logged regardless of status code
			logOutput := logBuffer.String()
			if !strings.Contains(logOutput, "GET") {
				t.Errorf("Expected request to be logged for status %d", statusCode)
			}
		}
	})
}

func TestJSONMiddleware(t *testing.T) {
	t.Run("sets JSON content type", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "test"}`))
		})

		wrappedHandler := jsonMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Check that Content-Type header is set
		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Check response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		if rr.Body.String() != `{"message": "test"}` {
			t.Errorf("Expected JSON response, got '%s'", rr.Body.String())
		}
	})

	t.Run("preserves other headers", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "custom-value")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := jsonMiddleware(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Check that JSON content type is set
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be set to application/json")
		}

		// Check that other headers are preserved
		if rr.Header().Get("X-Custom-Header") != "custom-value" {
			t.Error("Expected custom header to be preserved")
		}

		if rr.Header().Get("Cache-Control") != "no-cache" {
			t.Error("Expected cache control header to be preserved")
		}
	})

	t.Run("works with different HTTP methods", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123"}`))
		})

		wrappedHandler := jsonMiddleware(testHandler)

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type 'application/json' for %s method", method)
			}
		}
	})

	t.Run("handles empty response", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			// No body written
		})

		wrappedHandler := jsonMiddleware(testHandler)

		req := httptest.NewRequest("DELETE", "/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Should still set Content-Type even for empty responses
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be set even for empty responses")
		}

		if rr.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", rr.Code)
		}

		if rr.Body.String() != "" {
			t.Errorf("Expected empty body, got '%s'", rr.Body.String())
		}
	})
}

func TestMiddlewareChaining(t *testing.T) {
	t.Run("chains middleware correctly", func(t *testing.T) {
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		})

		// Chain both middleware
		wrappedHandler := loggingMiddleware(jsonMiddleware(testHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:8080"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Check that both middleware effects are present
		// JSON middleware should set Content-Type
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected JSON middleware to set Content-Type")
		}

		// Logging middleware should log the request
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "GET") || !strings.Contains(logOutput, "/test") {
			t.Error("Expected logging middleware to log the request")
		}

		// Response should be correct
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		if rr.Body.String() != `{"status": "ok"}` {
			t.Errorf("Expected JSON response, got '%s'", rr.Body.String())
		}
	})

	t.Run("middleware order independence", func(t *testing.T) {
		var logBuffer bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&logBuffer)
		defer log.SetOutput(originalOutput)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Test different order: JSON then logging
		wrappedHandler := jsonMiddleware(loggingMiddleware(testHandler))

		req := httptest.NewRequest("POST", "/api/test", nil)
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		// Both effects should still be present regardless of order
		if rr.Header().Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type to be set regardless of middleware order")
		}

		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, "POST") {
			t.Error("Expected request to be logged regardless of middleware order")
		}
	})
}
