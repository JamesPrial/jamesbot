package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Server provides an HTTP API for controlling and querying the bot.
// It listens only on localhost (127.0.0.1) for security.
type Server struct {
	port       int
	bot        BotInfo
	logger     zerolog.Logger
	httpServer *http.Server
	listener   net.Listener
}

// NewServer creates a new control API server.
// The server will bind to 127.0.0.1:port when started.
func NewServer(port int, bot BotInfo, logger zerolog.Logger) *Server {
	s := &Server{
		port:   port,
		bot:    bot,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/rules", s.handleRules)
	mux.HandleFunc("/rules/set", s.handleSetRule)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s
}

// Start starts the HTTP server on localhost.
// Returns an error if the server fails to start.
func (s *Server) Start() error {
	if s == nil {
		return fmt.Errorf("server cannot be nil")
	}

	if s.listener != nil {
		return errors.New("server already started")
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.httpServer.Addr, err)
	}

	s.listener = listener

	s.logger.Info().
		Str("address", listener.Addr().String()).
		Msg("control API server starting")

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("control API server error")
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server.
// The provided context can be used to set a deadline for the shutdown process.
func (s *Server) Stop(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("server cannot be nil")
	}

	// Return early if server was never started
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info().Msg("stopping control API server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown control API server: %w", err)
	}

	s.logger.Info().Msg("control API server stopped")
	return nil
}

// Addr returns the listener address of the server.
// Returns empty string if the server is not started.
func (s *Server) Addr() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// ServeHTTP implements http.Handler interface, allowing the server to be used with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}

// handleStats handles GET /stats requests.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.bot.Stats()
	if stats == nil {
		s.logger.Error().Msg("bot returned nil stats")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error().Err(err).Msg("failed to encode stats")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleRules handles GET /rules requests.
func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rules := s.bot.Rules()
	if rules == nil {
		rules = []Rule{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rules); err != nil {
		s.logger.Error().Err(err).Msg("failed to encode rules")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// SetRuleRequest represents the JSON payload for setting a rule.
type SetRuleRequest struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// handleSetRule handles POST /rules/set requests.
func (s *Server) handleSetRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn().Err(err).Msg("invalid request body")
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Key == "" {
		http.Error(w, "Bad request: name and key are required", http.StatusBadRequest)
		return
	}

	if err := s.bot.SetRule(req.Name, req.Key, req.Value); err != nil {
		s.logger.Error().
			Err(err).
			Str("name", req.Name).
			Str("key", req.Key).
			Msg("failed to set rule")

		// Return 400 for rule not found, 500 for other errors
		statusCode := http.StatusInternalServerError
		if err == ErrRuleNotFound {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, fmt.Sprintf("Failed to set rule: %v", err), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error().Err(err).Msg("failed to encode response")
	}
}
