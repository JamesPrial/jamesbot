package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"jamesbot/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// createMockServer creates an httptest.Server with the given handler function.
func createMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// statsResponse returns a valid JSON response for the /stats endpoint.
func statsResponse() string {
	return `{
		"uptime": "5m0s",
		"start_time": 1704067200,
		"commands_executed": 42,
		"guild_count": 3,
		"active_rules": 2
	}`
}

// rulesResponse returns a valid JSON array response for the /rules endpoint.
func rulesResponse() string {
	return `[
		{"name": "spam-filter", "description": "Filters spam messages", "enabled": true, "key": "threshold", "value": "10"},
		{"name": "link-filter", "description": "Filters unwanted links", "enabled": false, "key": "domains", "value": "example.com"}
	]`
}

// emptyRulesResponse returns an empty JSON array for the /rules endpoint.
func emptyRulesResponse() string {
	return `[]`
}

// =============================================================================
// NewClient Tests
// =============================================================================

func Test_NewClient_ValidEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "valid endpoint returns non-nil Client",
			endpoint: "http://127.0.0.1:8765",
		},
		{
			name:     "localhost with different port",
			endpoint: "http://localhost:9000",
		},
		{
			name:     "endpoint with trailing slash",
			endpoint: "http://127.0.0.1:8765/",
		},
		{
			name:     "https endpoint",
			endpoint: "https://api.example.com:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := api.NewClient(tt.endpoint)

			require.NotNil(t, client, "NewClient should return non-nil Client")
		})
	}
}

func Test_NewClient_AnyEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "standard endpoint is set",
			endpoint: "http://127.0.0.1:8765",
		},
		{
			name:     "empty string endpoint",
			endpoint: "",
		},
		{
			name:     "malformed URL",
			endpoint: "not-a-valid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := api.NewClient(tt.endpoint)

			require.NotNil(t, client, "NewClient should return non-nil Client for any input")
		})
	}
}

func Test_NewClient_Has10SecondTimeout(t *testing.T) {
	client := api.NewClient("http://127.0.0.1:8765")

	require.NotNil(t, client, "NewClient should return non-nil Client")

	// The client should have a 10 second timeout configured
	// We verify this by checking the client's internal HTTP client timeout
	timeout := client.Timeout()
	assert.Equal(t, 10*time.Second, timeout, "Client should have 10 second timeout")
}

// =============================================================================
// NewClient Pre-computed URL Tests
// =============================================================================

func Test_NewClient_PrecomputedURLs(t *testing.T) {
	tests := []struct {
		name             string
		endpoint         string
		wantStatsPath    string
		wantRulesPath    string
		wantRulesSetPath string
		description      string
	}{
		{
			name:             "URLs pre-computed correctly",
			endpoint:         "http://localhost:8080",
			wantStatsPath:    "/stats",
			wantRulesPath:    "/rules",
			wantRulesSetPath: "/rules/set",
			description:      "standard endpoint without trailing slash",
		},
		{
			name:             "URLs work with trailing slash",
			endpoint:         "http://localhost:8080/",
			wantStatsPath:    "/stats",
			wantRulesPath:    "/rules",
			wantRulesSetPath: "/rules/set",
			description:      "endpoint with trailing slash should not cause double slash",
		},
		{
			name:             "empty endpoint",
			endpoint:         "",
			wantStatsPath:    "/stats",
			wantRulesPath:    "/rules",
			wantRulesSetPath: "/rules/set",
			description:      "empty endpoint should produce paths starting with /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track the actual paths received by the server
			var statsPathReceived string
			var rulesPathReceived string
			var rulesSetPathReceived string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				switch {
				case strings.HasSuffix(path, "/stats") || path == "/stats":
					statsPathReceived = path
					_, _ = w.Write([]byte(statsResponse()))
				case strings.HasSuffix(path, "/rules/set") || path == "/rules/set":
					rulesSetPathReceived = path
					_, _ = w.Write([]byte(`{"status": "ok"}`))
				case strings.HasSuffix(path, "/rules") || path == "/rules":
					rulesPathReceived = path
					_, _ = w.Write([]byte(emptyRulesResponse()))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// For non-empty endpoints, use the test server URL
			// For empty endpoint test, we cannot actually make requests
			if tt.endpoint == "" {
				// For empty endpoint, just verify the client is created
				// The URLs will be "/stats", "/rules", "/rules/set"
				client := api.NewClient(tt.endpoint)
				require.NotNil(t, client, "NewClient should return non-nil Client for empty endpoint")
				// We cannot verify network behavior with empty endpoint
				// but the client should be created with the expected path suffixes
				return
			}

			// Replace the endpoint with our test server URL, preserving trailing slash behavior
			var testEndpoint string
			if strings.HasSuffix(tt.endpoint, "/") {
				testEndpoint = server.URL + "/"
			} else {
				testEndpoint = server.URL
			}

			client := api.NewClient(testEndpoint)
			require.NotNil(t, client, "NewClient should return non-nil Client")

			// Test GetStats path
			_, err := client.GetStats()
			require.NoError(t, err, "GetStats should succeed")
			assert.Equal(t, tt.wantStatsPath, statsPathReceived,
				"statsURL should produce correct path: %s", tt.description)

			// Test ListRules path
			_, err = client.ListRules()
			require.NoError(t, err, "ListRules should succeed")
			assert.Equal(t, tt.wantRulesPath, rulesPathReceived,
				"rulesURL should produce correct path: %s", tt.description)

			// Test SetRule path
			err = client.SetRule("test", "key", "value")
			require.NoError(t, err, "SetRule should succeed")
			assert.Equal(t, tt.wantRulesSetPath, rulesSetPathReceived,
				"rulesSetURL should produce correct path: %s", tt.description)
		})
	}
}

func Test_NewClient_PrecomputedURLs_NoDoubleSlash(t *testing.T) {
	// This test specifically verifies that trailing slashes in the endpoint
	// do not result in double slashes (e.g., "http://localhost:8080//stats")

	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	}))
	defer server.Close()

	// Create client with trailing slash in endpoint
	client := api.NewClient(server.URL + "/")

	_, err := client.GetStats()
	require.NoError(t, err, "GetStats should succeed")

	// Verify no double slash in the path
	assert.NotContains(t, receivedPath, "//",
		"URL path should not contain double slashes")
	assert.Equal(t, "/stats", receivedPath,
		"path should be /stats without leading double slash")
}

func Test_NewClient_PrecomputedURLs_EmptyEndpoint(t *testing.T) {
	// Verify client creation with empty endpoint
	client := api.NewClient("")
	require.NotNil(t, client, "NewClient should return non-nil Client for empty endpoint")

	// The client is created, but network requests will fail
	// because the URLs will be just "/stats", "/rules", "/rules/set"
	// which are not valid URLs for HTTP requests

	// Verify timeout is still set correctly
	assert.Equal(t, 10*time.Second, client.Timeout(),
		"Client with empty endpoint should still have 10 second timeout")
}

// =============================================================================
// GetStats Tests
// =============================================================================

func Test_GetStats_SuccessfulRequest(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/stats", r.URL.Path, "request path should be /stats")
		assert.Equal(t, http.MethodGet, r.Method, "request method should be GET")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	stats, err := client.GetStats()

	require.NoError(t, err, "GetStats should not return error on successful request")
	require.NotNil(t, stats, "GetStats should return non-nil Stats")

	assert.Equal(t, "5m0s", stats.Uptime)
	assert.Equal(t, int64(1704067200), stats.StartTime)
	assert.Equal(t, int64(42), stats.CommandsExecuted)
	assert.Equal(t, 3, stats.GuildCount)
	assert.Equal(t, 2, stats.ActiveRules)
}

func Test_GetStats_ServerDown(t *testing.T) {
	// Use an endpoint where no server is running
	client := api.NewClient("http://127.0.0.1:59999")

	stats, err := client.GetStats()

	require.Error(t, err, "GetStats should return error when server is down")
	assert.Nil(t, stats, "GetStats should return nil Stats when server is down")
	assert.Contains(t, strings.ToLower(err.Error()), "connection",
		"error should contain 'connection'")
}

func Test_GetStats_Non200Response(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "server returns 500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "server returns 404 Not Found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "server returns 503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
		},
		{
			name:       "server returns 400 Bad Request",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte("error"))
			})
			defer server.Close()

			client := api.NewClient(server.URL)

			stats, err := client.GetStats()

			require.Error(t, err, "GetStats should return error for non-200 response")
			assert.Nil(t, stats, "GetStats should return nil Stats for non-200 response")
			assert.Contains(t, strings.ToLower(err.Error()), "unexpected status",
				"error should contain 'unexpected status'")
		})
	}
}

func Test_GetStats_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{
			name:     "malformed JSON with missing brace",
			response: `{"uptime": "5m0s"`,
		},
		{
			name:     "completely invalid JSON",
			response: `not json at all`,
		},
		{
			name:     "empty response body",
			response: ``,
		},
		{
			name:     "JSON array instead of object",
			response: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			})
			defer server.Close()

			client := api.NewClient(server.URL)

			stats, err := client.GetStats()

			require.Error(t, err, "GetStats should return error for invalid JSON")
			assert.Nil(t, stats, "GetStats should return nil Stats for invalid JSON")
			assert.Contains(t, strings.ToLower(err.Error()), "decode",
				"error should contain 'decode'")
		})
	}
}

// =============================================================================
// ListRules Tests
// =============================================================================

func Test_ListRules_SuccessfulRequest(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rules", r.URL.Path, "request path should be /rules")
		assert.Equal(t, http.MethodGet, r.Method, "request method should be GET")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(rulesResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	rules, err := client.ListRules()

	require.NoError(t, err, "ListRules should not return error on successful request")
	require.NotNil(t, rules, "ListRules should return non-nil slice")

	assert.Len(t, rules, 2, "ListRules should return 2 rules")
	assert.Equal(t, "spam-filter", rules[0].Name)
	assert.Equal(t, "Filters spam messages", rules[0].Description)
	assert.True(t, rules[0].Enabled)
}

func Test_ListRules_EmptyRules(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(emptyRulesResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	rules, err := client.ListRules()

	require.NoError(t, err, "ListRules should not return error for empty rules")
	require.NotNil(t, rules, "ListRules should return non-nil slice")
	assert.Empty(t, rules, "ListRules should return empty slice")
}

func Test_ListRules_ServerDown(t *testing.T) {
	// Use an endpoint where no server is running
	client := api.NewClient("http://127.0.0.1:59998")

	rules, err := client.ListRules()

	require.Error(t, err, "ListRules should return error when server is down")
	assert.Nil(t, rules, "ListRules should return nil slice when server is down")
	assert.Contains(t, strings.ToLower(err.Error()), "connection",
		"error should contain 'connection'")
}

func Test_ListRules_Non200Response(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "server returns 404 Not Found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "server returns 500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "server returns 403 Forbidden",
			statusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte("error"))
			})
			defer server.Close()

			client := api.NewClient(server.URL)

			rules, err := client.ListRules()

			require.Error(t, err, "ListRules should return error for non-200 response")
			assert.Nil(t, rules, "ListRules should return nil slice for non-200 response")
			assert.Contains(t, strings.ToLower(err.Error()), "unexpected status",
				"error should contain 'unexpected status'")
		})
	}
}

func Test_ListRules_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{
			name:     "malformed JSON array",
			response: `[{"name": "test"`,
		},
		{
			name:     "object instead of array",
			response: `{"name": "test"}`,
		},
		{
			name:     "invalid JSON content",
			response: `not json`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			})
			defer server.Close()

			client := api.NewClient(server.URL)

			rules, err := client.ListRules()

			require.Error(t, err, "ListRules should return error for invalid JSON")
			assert.Nil(t, rules, "ListRules should return nil slice for invalid JSON")
			assert.Contains(t, strings.ToLower(err.Error()), "decode",
				"error should contain 'decode'")
		})
	}
}

// =============================================================================
// SetRule Tests
// =============================================================================

func Test_SetRule_SuccessfulUpdate(t *testing.T) {
	var receivedRequest struct {
		Name  string `json:"name"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rules/set", r.URL.Path, "request path should be /rules/set")
		assert.Equal(t, http.MethodPost, r.Method, "request method should be POST")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		err = json.Unmarshal(body, &receivedRequest)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	err := client.SetRule("spam-filter", "threshold", "10")

	require.NoError(t, err, "SetRule should not return error on successful update")
	assert.Equal(t, "spam-filter", receivedRequest.Name)
	assert.Equal(t, "threshold", receivedRequest.Key)
	assert.Equal(t, "10", receivedRequest.Value)
}

func Test_SetRule_ServerDown(t *testing.T) {
	// Use an endpoint where no server is running
	client := api.NewClient("http://127.0.0.1:59997")

	err := client.SetRule("spam-filter", "threshold", "10")

	require.Error(t, err, "SetRule should return error when server is down")
	assert.Contains(t, strings.ToLower(err.Error()), "connection",
		"error should contain 'connection'")
}

func Test_SetRule_ServerReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "server returns 400 Bad Request",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "server returns 500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "server returns 404 Not Found",
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte("error occurred"))
			})
			defer server.Close()

			client := api.NewClient(server.URL)

			err := client.SetRule("spam-filter", "threshold", "10")

			require.Error(t, err, "SetRule should return error when server returns error status")
			assert.Contains(t, strings.ToLower(err.Error()), "rule update failed",
				"error should contain 'rule update failed'")
		})
	}
}

func Test_SetRule_WithEmptyValue(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	err := client.SetRule("spam-filter", "threshold", "")

	require.NoError(t, err, "SetRule should allow empty value")
}

func Test_SetRule_WithSpecialCharacters(t *testing.T) {
	var receivedRequest struct {
		Name  string `json:"name"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedRequest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	err := client.SetRule("spam-filter", "message", "Hello, World! @#$%^&*()")

	require.NoError(t, err, "SetRule should handle special characters")
	assert.Equal(t, "Hello, World! @#$%^&*()", receivedRequest.Value)
}

func Test_SetRule_WithUnicode(t *testing.T) {
	var receivedRequest struct {
		Name  string `json:"name"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedRequest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	err := client.SetRule("spam-filter", "message", "Hello, World!")

	require.NoError(t, err, "SetRule should handle unicode characters")
	assert.Contains(t, receivedRequest.Value, "Hello")
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func Test_Client_EndpointWithTrailingSlash(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	})
	defer server.Close()

	// Client with trailing slash in endpoint
	client := api.NewClient(server.URL + "/")

	stats, err := client.GetStats()

	require.NoError(t, err, "GetStats should handle endpoint with trailing slash")
	require.NotNil(t, stats)
}

func Test_Client_SlowServer(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		// Delay response by 100ms (within timeout)
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	stats, err := client.GetStats()

	require.NoError(t, err, "GetStats should succeed for slow but responsive server")
	require.NotNil(t, stats)
}

func Test_GetStats_ZeroValues(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"uptime": "0s",
			"start_time": 0,
			"commands_executed": 0,
			"guild_count": 0,
			"active_rules": 0
		}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	stats, err := client.GetStats()

	require.NoError(t, err, "GetStats should handle zero values")
	require.NotNil(t, stats)
	assert.Equal(t, "0s", stats.Uptime)
	assert.Equal(t, int64(0), stats.CommandsExecuted)
	assert.Equal(t, 0, stats.GuildCount)
}

func Test_ListRules_SingleRule(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name": "only-rule", "description": "The only rule", "enabled": true, "key": "", "value": ""}]`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	rules, err := client.ListRules()

	require.NoError(t, err, "ListRules should handle single rule")
	require.NotNil(t, rules)
	assert.Len(t, rules, 1)
	assert.Equal(t, "only-rule", rules[0].Name)
}

func Test_ListRules_ManyRules(t *testing.T) {
	// Create a response with 100 rules
	var rules []map[string]interface{}
	for i := 0; i < 100; i++ {
		rules = append(rules, map[string]interface{}{
			"name":        "rule-" + string(rune('0'+i%10)),
			"description": "Rule description",
			"enabled":     i%2 == 0,
			"key":         "key",
			"value":       "value",
		})
	}
	rulesJSON, _ := json.Marshal(rules)

	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rulesJSON)
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	result, err := client.ListRules()

	require.NoError(t, err, "ListRules should handle many rules")
	require.NotNil(t, result)
	assert.Len(t, result, 100)
}

// =============================================================================
// Concurrent Request Tests
// =============================================================================

func Test_Client_ConcurrentGetStats(t *testing.T) {
	var requestCount int32
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	numRequests := 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.GetStats()
			results <- err
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount, "all concurrent requests should succeed")
}

// =============================================================================
// Request Header Tests
// =============================================================================

func Test_SetRule_SendsCorrectContentType(t *testing.T) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json",
			"Content-Type header should be application/json")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	err := client.SetRule("test", "key", "value")
	require.NoError(t, err)
}

// =============================================================================
// Nil Safety Tests
// =============================================================================

func Test_GetStats_NilClient(t *testing.T) {
	var client *api.Client = nil

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetStats on nil client should not panic: %v", r)
		}
	}()

	stats, err := client.GetStats()
	assert.Error(t, err, "GetStats on nil client should return error")
	assert.Nil(t, stats, "GetStats on nil client should return nil stats")
}

func Test_ListRules_NilClient(t *testing.T) {
	var client *api.Client = nil

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ListRules on nil client should not panic: %v", r)
		}
	}()

	rules, err := client.ListRules()
	assert.Error(t, err, "ListRules on nil client should return error")
	assert.Nil(t, rules, "ListRules on nil client should return nil rules")
}

func Test_SetRule_NilClient(t *testing.T) {
	var client *api.Client = nil

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetRule on nil client should not panic: %v", r)
		}
	}()

	err := client.SetRule("name", "key", "value")
	assert.Error(t, err, "SetRule on nil client should return error")
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func Benchmark_NewClient(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = api.NewClient("http://127.0.0.1:8765")
	}
}

func Benchmark_GetStats(b *testing.B) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(statsResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.GetStats()
	}
}

func Benchmark_ListRules(b *testing.B) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(rulesResponse()))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ListRules()
	}
}

func Benchmark_SetRule(b *testing.B) {
	server := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	client := api.NewClient(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SetRule("spam-filter", "threshold", "10")
	}
}
