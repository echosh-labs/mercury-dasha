package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const maxRequestSize = 10 << 20 // 10 MB

// Server provides HTTP and Stdio transports for Mercury Dasha MCP.
type Server struct {
	handler *Handler
	apiKey  string
}

// NewServer creates an MCP server with the given handler and optional API key.
func NewServer(handler *Handler, apiKey string) *Server {
	return &Server{
		handler: handler,
		apiKey:  apiKey,
	}
}

// RegisterRoutes registers the Streamable HTTP MCP endpoint on the provided mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/mcp", s.handleMCP)
}

// handleMCP serves the Streamable HTTP transport endpoint for MCP JSON-RPC.
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for web-based MCP inspectors and remote LLM agents
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service":          "mercury-dasha-mcp",
			"protocol_version": ProtocolVersion,
			"transport":        "streamable-http",
			"endpoint":         "/mcp",
			"description":      "Mercury Dasha Sovereign Scribe & Chrono-Matrix MCP Server",
		})
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, GET, OPTIONS")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestSize))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	resp := s.handler.HandleRequest(r.Context(), body)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[MCP HTTP] Transport error encoding response: %v", err)
	}
}

func (s *Server) authenticate(r *http.Request) bool {
	if s.apiKey == "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	return token == s.apiKey
}

// ServeStdio runs the standard input/output transport for CLI or local IDE integration.
func (s *Server) ServeStdio(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	// Allow up to 10MB per line
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxRequestSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		resp := s.handler.HandleRequest(context.Background(), line)
		if resp == nil {
			continue
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}

		if _, err := fmt.Fprintf(out, "%s\n", string(respBytes)); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	return scanner.Err()
}
