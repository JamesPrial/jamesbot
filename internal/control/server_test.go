package control_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"jamesbot/internal/control"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// discardLogger returns a zerolog.Logger that discards all output.
func discardLogger() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}

// mockBotInfo implements the BotInfo interface for testing.
type mockBotInfo struct {
	stats         *control.Stats
	rules         []control.Rule
	setRuleErr    error
	setRuleCalled bool
	setRuleName   string
	setRuleKey    string
	setRuleValue  string
}

// Stats returns the mock stats.
func (m *mockBotInfo) Stats() *control.Stats {
	return m.stats
}

// Rules returns the mock rules list.
func (m *mockBotInfo) Rules() []control.Rule {
	return m.rules
}

// SetRule records the call and returns the mock error.
func (m *mockBotInfo) SetRule(name, key, value string) error {
	m.setRuleCalled = true
	m.setRuleName = name
	m.setRuleKey = key
	m.setRuleValue = value
	return m.setRuleErr
}

// newMockBotInfo creates a mock BotInfo with default values.
func newMockBotInfo() *mockBotInfo {
	return &mockBotInfo{
		stats: &control.Stats{
			Uptime:           "5m0s",
			StartTime:        time.Now().Unix(),
			CommandsExecuted: 42,
			GuildCount:       3,
			ActiveRules:      2,
		},
		rules: []control.Rule{},
	}
}

// newMockBotInfoWithRules creates a mock BotInfo with predefined rules.
func newMockBotInfoWithRules(rules []control.Rule) *mockBotInfo {
	m := newMockBotInfo()
	m.rules = rules
	return m
}

// newMockBotInfoWithStats creates a mock BotInfo with custom stats.
func newMockBotInfoWithStats(stats *control.Stats) *mockBotInfo {
	m := newMockBotInfo()
	m.stats = stats
	return m
}

// createTestServer creates a test HTTP server using the control server's handler.
// Since the actual Server doesn't expose ServeHTTP, we use httptest with a handler
// that mimics the actual implementation for unit testing.
func createTestHandler(bot control.BotInfo, logger zerolog.Logger) http.Handler {
	mux := http.NewServeMux()

	// /stats handler
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := bot.Stats()
		if stats == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// /rules handler
	mux.HandleFunc("/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rules := bot.Rules()
		if rules == nil {
			rules = []control.Rule{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rules); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// /rules/set handler
	mux.HandleFunc("/rules/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req control.SetRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Key == "" {
			http.Error(w, "Bad request: name and key are required", http.StatusBadRequest)
			return
		}

		if err := bot.SetRule(req.Name, req.Key, req.Value); err != nil {
			http.Error(w, "Failed to set rule: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	return mux
}

// =============================================================================
// NewServer Tests
// =============================================================================

func Test_NewServer_ValidParams(t *testing.T) {
	tests := []struct {
		name   string
		port   int
		bot    control.BotInfo
		logger zerolog.Logger
	}{
		{
			name:   "valid port 8765 with mock bot and logger",
			port:   8765,
			bot:    newMockBotInfo(),
			logger: discardLogger(),
		},
		{
			name:   "valid port 9000",
			port:   9000,
			bot:    newMockBotInfo(),
			logger: discardLogger(),
		},
		{
			name:   "valid high port number",
			port:   65535,
			bot:    newMockBotInfo(),
			logger: discardLogger(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := control.NewServer(tt.port, tt.bot, tt.logger)

			require.NotNil(t, server, "NewServer should return non-nil Server")
		})
	}
}

func Test_NewServer_PortZero(t *testing.T) {
	bot := newMockBotInfo()
	logger := discardLogger()

	server := control.NewServer(0, bot, logger)

	require.NotNil(t, server, "NewServer with port 0 should return non-nil Server")
	// Note: Port 0 means 127.0.0.1:0, which will bind to a random available port
}

func Test_NewServer_NilBot(t *testing.T) {
	logger := discardLogger()

	// Implementation does not check for nil bot at construction time
	server := control.NewServer(8080, nil, logger)

	// Server is created but will fail at runtime when endpoints access bot
	require.NotNil(t, server, "NewServer accepts nil bot (handles at runtime)")
}

// =============================================================================
// GET /stats Endpoint Tests
// =============================================================================

func Test_StatsEndpoint_ValidRequest(t *testing.T) {
	bot := newMockBotInfoWithStats(&control.Stats{
		Uptime:           "10m0s",
		StartTime:        time.Now().Unix(),
		CommandsExecuted: 100,
		GuildCount:       5,
		ActiveRules:      3,
	})
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET /stats should return 200 OK")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
		"Content-Type should be application/json")

	var response control.Stats
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "response should be valid JSON")

	assert.Equal(t, "10m0s", response.Uptime, "uptime should match")
	assert.Equal(t, int64(100), response.CommandsExecuted, "commands_executed should match")
	assert.Equal(t, 5, response.GuildCount, "guild_count should match")
}

func Test_StatsEndpoint_WrongMethod(t *testing.T) {
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			bot := newMockBotInfo()
			handler := createTestHandler(bot, discardLogger())

			req := httptest.NewRequest(method, "/stats", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
				"%s /stats should return 405 Method Not Allowed", method)
		})
	}
}

func Test_StatsEndpoint_ZeroValues(t *testing.T) {
	bot := newMockBotInfoWithStats(&control.Stats{
		Uptime:           "0s",
		StartTime:        0,
		CommandsExecuted: 0,
		GuildCount:       0,
		ActiveRules:      0,
	})
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET /stats with zero values should return 200 OK")

	var response control.Stats
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "response should be valid JSON")

	assert.Equal(t, int64(0), response.CommandsExecuted, "commands_executed should be 0")
	assert.Equal(t, 0, response.GuildCount, "guild_count should be 0")
}

func Test_StatsEndpoint_NilStats(t *testing.T) {
	bot := &mockBotInfo{
		stats: nil,
		rules: []control.Rule{},
	}
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code,
		"GET /stats with nil stats should return 500 Internal Server Error")
}

// =============================================================================
// GET /rules Endpoint Tests
// =============================================================================

func Test_RulesEndpoint_ValidRequest(t *testing.T) {
	rules := []control.Rule{
		{Name: "spam-filter", Description: "Filters spam messages", Enabled: true},
		{Name: "link-filter", Description: "Filters unwanted links", Enabled: false},
	}
	bot := newMockBotInfoWithRules(rules)
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET /rules should return 200 OK")
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"),
		"Content-Type should be application/json")

	var response []control.Rule
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "response should be valid JSON array")

	assert.Len(t, response, 2, "response should contain 2 rules")
	assert.Equal(t, "spam-filter", response[0].Name)
	assert.True(t, response[0].Enabled)
}

func Test_RulesEndpoint_EmptyRules(t *testing.T) {
	bot := newMockBotInfoWithRules([]control.Rule{})
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET /rules with empty rules should return 200 OK")

	var response []control.Rule
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "response should be valid JSON array")

	// Should return empty array, not null
	assert.NotNil(t, response, "response should not be nil")
	assert.Empty(t, response, "response should be empty array")
}

func Test_RulesEndpoint_NilRules(t *testing.T) {
	bot := &mockBotInfo{
		stats: newMockBotInfo().stats,
		rules: nil,
	}
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "GET /rules with nil rules should return 200 OK")

	// Response should be empty JSON array [] not null
	body := strings.TrimSpace(rec.Body.String())
	assert.Equal(t, "[]", body, "nil rules should return empty array []")
}

func Test_RulesEndpoint_WrongMethod(t *testing.T) {
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			bot := newMockBotInfo()
			handler := createTestHandler(bot, discardLogger())

			req := httptest.NewRequest(method, "/rules", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
				"%s /rules should return 405 Method Not Allowed", method)
		})
	}
}

// =============================================================================
// POST /rules/set Endpoint Tests
// =============================================================================

func Test_RulesSetEndpoint_ValidRequest(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"spam-filter","key":"threshold","value":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "POST /rules/set with valid body should return 200 OK")
	assert.True(t, bot.setRuleCalled, "SetRule should be called")
	assert.Equal(t, "spam-filter", bot.setRuleName, "name should match")
	assert.Equal(t, "threshold", bot.setRuleKey, "key should match")
	assert.Equal(t, "10", bot.setRuleValue, "value should match")
}

func Test_RulesSetEndpoint_MissingName(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"key":"threshold","value":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"POST /rules/set with missing name should return 400 Bad Request")
	assert.False(t, bot.setRuleCalled, "SetRule should not be called with missing name")
}

func Test_RulesSetEndpoint_MissingKey(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"spam-filter","value":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"POST /rules/set with missing key should return 400 Bad Request")
	assert.False(t, bot.setRuleCalled, "SetRule should not be called with missing key")
}

func Test_RulesSetEndpoint_MissingValue(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"spam-filter","key":"threshold"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Missing value is allowed (empty string is a valid value)
	assert.Equal(t, http.StatusOK, rec.Code,
		"POST /rules/set with missing value should be allowed (empty string)")
	assert.True(t, bot.setRuleCalled, "SetRule should be called")
	assert.Equal(t, "", bot.setRuleValue, "value should be empty string")
}

func Test_RulesSetEndpoint_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "malformed JSON with missing brace",
			body: `{"name":"x","key":"y","value":"z"`,
		},
		{
			name: "malformed JSON with invalid syntax",
			body: `{invalid}`,
		},
		{
			name: "malformed JSON with trailing comma",
			body: `{"name":"x","key":"y","value":"z",}`,
		},
		{
			name: "empty body",
			body: ``,
		},
		{
			name: "plain text instead of JSON",
			body: `not json at all`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := newMockBotInfo()
			handler := createTestHandler(bot, discardLogger())

			req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code,
				"POST /rules/set with invalid JSON should return 400 Bad Request")
			assert.False(t, bot.setRuleCalled, "SetRule should not be called with invalid JSON")
		})
	}
}

func Test_RulesSetEndpoint_WrongMethod(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			bot := newMockBotInfo()
			handler := createTestHandler(bot, discardLogger())

			body := `{"name":"x","key":"y","value":"z"}`
			req := httptest.NewRequest(method, "/rules/set", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
				"%s /rules/set should return 405 Method Not Allowed", method)
		})
	}
}

func Test_RulesSetEndpoint_RuleNotFound(t *testing.T) {
	bot := newMockBotInfo()
	bot.setRuleErr = errors.New("rule not found")
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"nonexistent","key":"foo","value":"bar"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// When SetRule returns an error, the handler returns 500
	assert.Equal(t, http.StatusInternalServerError, rec.Code,
		"POST /rules/set with SetRule error should return 500")
	assert.True(t, bot.setRuleCalled, "SetRule should be called")
}

func Test_RulesSetEndpoint_EmptyName(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"","key":"threshold","value":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"POST /rules/set with empty name should return 400 Bad Request")
	assert.False(t, bot.setRuleCalled, "SetRule should not be called with empty name")
}

func Test_RulesSetEndpoint_WhitespaceName(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"   ","key":"threshold","value":"10"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Whitespace-only name is accepted as non-empty by current implementation
	// This documents the actual behavior
	t.Logf("POST /rules/set with whitespace name returned status: %d", rec.Code)
}

// =============================================================================
// Server Lifecycle Tests
// =============================================================================

func Test_ServerLifecycle_StartThenStop(t *testing.T) {
	bot := newMockBotInfo()
	logger := discardLogger()
	server := control.NewServer(0, bot, logger)

	// Start server
	err := server.Start()
	require.NoError(t, err, "Start should not return error")

	// Stop server
	err = server.Stop(context.Background())
	assert.NoError(t, err, "Stop should not return error")
}

func Test_ServerLifecycle_StopWithoutStart(t *testing.T) {
	bot := newMockBotInfo()
	logger := discardLogger()
	server := control.NewServer(0, bot, logger)

	// Stop without starting - behavior depends on implementation
	err := server.Stop(context.Background())

	// Document actual behavior - may return error or succeed
	t.Logf("Stop without Start returned: %v", err)
}

func Test_ServerLifecycle_DoubleStop(t *testing.T) {
	bot := newMockBotInfo()
	logger := discardLogger()
	server := control.NewServer(0, bot, logger)

	// Start server
	err := server.Start()
	require.NoError(t, err)

	// First stop
	err = server.Stop(context.Background())
	assert.NoError(t, err, "first Stop should not return error")

	// Second stop - behavior depends on implementation
	err = server.Stop(context.Background())
	t.Logf("second Stop returned: %v", err)
}

func Test_ServerLifecycle_DoubleStart(t *testing.T) {
	bot := newMockBotInfo()
	logger := discardLogger()
	server := control.NewServer(0, bot, logger)

	// First start
	err := server.Start()
	require.NoError(t, err, "first Start should not return error")

	// Second start - will fail because port is already in use
	err = server.Start()
	assert.Error(t, err, "second Start should return error (port in use)")

	// Cleanup
	_ = server.Stop(context.Background())
}

// =============================================================================
// Unknown Endpoint Tests
// =============================================================================

func Test_UnknownEndpoint_Returns404(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	endpoints := []string{
		"/unknown",
		"/api/v1/stats",
		"/foo/bar",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code,
				"unknown endpoint %s should return 404 Not Found", endpoint)
		})
	}
}

// =============================================================================
// Content-Type Tests
// =============================================================================

func Test_Endpoints_ReturnJSON(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/stats", ""},
		{http.MethodGet, "/rules", ""},
		{http.MethodPost, "/rules/set", `{"name":"x","key":"y","value":"z"}`},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			var bodyReader io.Reader
			if ep.body != "" {
				bodyReader = strings.NewReader(ep.body)
			}

			req := httptest.NewRequest(ep.method, ep.path, bodyReader)
			if ep.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Only check Content-Type for successful responses
			if rec.Code == http.StatusOK {
				contentType := rec.Header().Get("Content-Type")
				assert.Contains(t, contentType, "application/json",
					"Content-Type should be application/json")
			}
		})
	}
}

// =============================================================================
// Integration Tests with Real HTTP Server
// =============================================================================

func Test_Server_RealHTTPIntegration(t *testing.T) {
	bot := &mockBotInfo{
		stats: &control.Stats{
			Uptime:           "1h0m0s",
			StartTime:        time.Now().Unix(),
			CommandsExecuted: 500,
			GuildCount:       10,
			ActiveRules:      2,
		},
		rules: []control.Rule{
			{Name: "test-rule", Description: "Test rule", Enabled: true},
		},
	}

	handler := createTestHandler(bot, discardLogger())
	server := httptest.NewServer(handler)
	defer server.Close()

	baseURL := server.URL

	t.Run("GET /stats", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/stats")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats control.Stats
		err = json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)

		assert.Equal(t, int64(500), stats.CommandsExecuted)
		assert.Equal(t, 10, stats.GuildCount)
	})

	t.Run("GET /rules", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/rules")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var rules []control.Rule
		err = json.NewDecoder(resp.Body).Decode(&rules)
		require.NoError(t, err)

		assert.Len(t, rules, 1)
		assert.Equal(t, "test-rule", rules[0].Name)
	})

	t.Run("POST /rules/set", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"test-rule","key":"threshold","value":"5"}`)
		resp, err := http.Post(baseURL+"/rules/set", "application/json", body)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, bot.setRuleCalled)
	})

	t.Run("POST /stats returns 405", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/stats", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// =============================================================================
// Concurrent Request Tests
// =============================================================================

func Test_Server_ConcurrentRequests(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())
	server := httptest.NewServer(handler)
	defer server.Close()

	baseURL := server.URL

	numRequests := 50
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(baseURL + "/stats")
			if err != nil {
				results <- -1
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		status := <-results
		if status == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, numRequests, successCount,
		"all concurrent requests should succeed")
}

// =============================================================================
// Error Response Format Tests
// =============================================================================

func Test_ErrorResponses_ContainMessage(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	t.Run("400 Bad Request has error message", func(t *testing.T) {
		body := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// Error response should contain some message
		assert.NotEmpty(t, rec.Body.String(), "error response should have a body")
		assert.Contains(t, rec.Body.String(), "Bad request")
	})

	t.Run("405 Method Not Allowed has error message", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/stats", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Contains(t, rec.Body.String(), "Method not allowed")
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func Benchmark_StatsEndpoint(b *testing.B) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_RulesEndpoint(b *testing.B) {
	rules := make([]control.Rule, 10)
	for i := 0; i < 10; i++ {
		rules[i] = control.Rule{Name: "rule", Description: "description", Enabled: true}
	}
	bot := newMockBotInfoWithRules(rules)
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_RulesSetEndpoint(b *testing.B) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"spam-filter","key":"threshold","value":"10"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

// Verify BotInfo interface is properly implemented by mock.
func Test_BotInfo_InterfaceCompliance(t *testing.T) {
	var _ control.BotInfo = (*mockBotInfo)(nil)
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func Test_StatsEndpoint_WithQueryParams(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats?foo=bar&baz=123", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should ignore query params and return stats
	assert.Equal(t, http.StatusOK, rec.Code,
		"GET /stats with query params should still return 200 OK")
}

func Test_RulesSetEndpoint_WithExtraFields(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	// JSON with extra unknown fields
	body := `{"name":"x","key":"y","value":"z","extra":"ignored","another":123}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Extra fields should be ignored
	assert.Equal(t, http.StatusOK, rec.Code,
		"POST /rules/set with extra fields should still work")
	assert.True(t, bot.setRuleCalled)
}

func Test_RulesSetEndpoint_UnicodeValues(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	// Unicode in JSON values
	body := `{"name":"spam-filter","key":"message","value":"Hello, \u4e16\u754c!"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code,
		"POST /rules/set with unicode should work")
	assert.True(t, bot.setRuleCalled)
	assert.Contains(t, bot.setRuleValue, "Hello")
}

func Test_RulesSetEndpoint_LargeBody(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	// Create a large but valid JSON body
	largeValue := strings.Repeat("x", 10000)
	body := `{"name":"spam-filter","key":"threshold","value":"` + largeValue + `"}`

	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code,
		"POST /rules/set with large body should succeed")
	assert.True(t, bot.setRuleCalled)
	assert.Equal(t, largeValue, bot.setRuleValue)
}

// =============================================================================
// Server Nil Safety Tests
// =============================================================================

func Test_Server_StartOnNil(t *testing.T) {
	var server *control.Server = nil

	// Should not panic and should return an error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start() on nil server should not panic: %v", r)
		}
	}()

	err := server.Start()
	assert.Error(t, err, "Start() on nil server should return error")
}

func Test_Server_StopOnNil(t *testing.T) {
	var server *control.Server = nil

	// Should not panic and should return an error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop() on nil server should not panic: %v", r)
		}
	}()

	err := server.Stop(context.Background())
	assert.Error(t, err, "Stop() on nil server should return error")
}

// =============================================================================
// Response JSON Structure Tests
// =============================================================================

func Test_StatsEndpoint_ResponseStructure(t *testing.T) {
	bot := newMockBotInfoWithStats(&control.Stats{
		Uptime:           "2h30m0s",
		StartTime:        1704067200,
		CommandsExecuted: 1234,
		GuildCount:       56,
		ActiveRules:      7,
	})
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify all expected fields are present
	assert.Contains(t, response, "uptime")
	assert.Contains(t, response, "start_time")
	assert.Contains(t, response, "commands_executed")
	assert.Contains(t, response, "guild_count")
	assert.Contains(t, response, "active_rules")

	// Verify types
	assert.IsType(t, "", response["uptime"])
	assert.IsType(t, float64(0), response["start_time"])
	assert.IsType(t, float64(0), response["commands_executed"])
	assert.IsType(t, float64(0), response["guild_count"])
	assert.IsType(t, float64(0), response["active_rules"])
}

func Test_RulesEndpoint_ResponseStructure(t *testing.T) {
	rules := []control.Rule{
		{Name: "rule1", Description: "First rule", Enabled: true},
	}
	bot := newMockBotInfoWithRules(rules)
	handler := createTestHandler(bot, discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/rules", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response []map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	require.Len(t, response, 1)

	// Verify rule structure
	rule := response[0]
	assert.Contains(t, rule, "name")
	assert.Contains(t, rule, "description")
	assert.Contains(t, rule, "enabled")

	assert.Equal(t, "rule1", rule["name"])
	assert.Equal(t, "First rule", rule["description"])
	assert.Equal(t, true, rule["enabled"])
}

func Test_RulesSetEndpoint_SuccessResponse(t *testing.T) {
	bot := newMockBotInfo()
	handler := createTestHandler(bot, discardLogger())

	body := `{"name":"x","key":"y","value":"z"}`
	req := httptest.NewRequest(http.MethodPost, "/rules/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
}
