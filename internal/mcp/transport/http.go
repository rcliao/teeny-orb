package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rcliao/teeny-orb/internal/mcp"
)

// HTTPTransport implements MCP transport over HTTP
type HTTPTransport struct {
	server     *http.Server
	handler    *HTTPHandler
	addr       string
	debug      bool
	shutdownCh chan struct{}
}

// HTTPHandler handles HTTP requests for MCP
type HTTPHandler struct {
	mcpServer MCPMessageHandler
	debug     bool
	mutex     sync.RWMutex
}

// MCPMessageHandler defines the interface for handling MCP messages
type MCPMessageHandler interface {
	HandleMessage(ctx context.Context, msg *mcp.Message) (*mcp.Message, error)
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(addr string, mcpServer MCPMessageHandler, debug bool) *HTTPTransport {
	handler := &HTTPHandler{
		mcpServer: mcpServer,
		debug:     debug,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", handler.handleMCP)
	mux.HandleFunc("/health", handler.handleHealth)
	mux.HandleFunc("/status", handler.handleStatus)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &HTTPTransport{
		server:     server,
		handler:    handler,
		addr:       addr,
		debug:      debug,
		shutdownCh: make(chan struct{}),
	}
}

// Start starts the HTTP server
func (h *HTTPTransport) Start(ctx context.Context) error {
	if h.debug {
		fmt.Fprintf(os.Stderr, "Starting MCP HTTP server on %s\n", h.addr)
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if h.debug {
				fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
			}
		}
	}()

	// Wait for context cancellation or shutdown
	select {
	case <-ctx.Done():
		return h.Shutdown()
	case <-h.shutdownCh:
		return nil
	}
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPTransport) Shutdown() error {
	if h.debug {
		fmt.Fprintln(os.Stderr, "Shutting down MCP HTTP server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.server.Shutdown(ctx)
	close(h.shutdownCh)
	return err
}

// Send is not used in HTTP transport (handled via HTTP responses)
func (h *HTTPTransport) Send(ctx context.Context, msg *mcp.Message) error {
	return fmt.Errorf("Send not supported in HTTP transport - use HTTP responses")
}

// Receive is not used in HTTP transport (handled via HTTP requests)
func (h *HTTPTransport) Receive(ctx context.Context) (*mcp.Message, error) {
	return nil, fmt.Errorf("Receive not supported in HTTP transport - use HTTP requests")
}

// Close closes the HTTP transport
func (h *HTTPTransport) Close() error {
	return h.Shutdown()
}

// HTTP Handler Methods

// handleMCP handles MCP JSON-RPC requests over HTTP
func (h *HTTPHandler) handleMCP(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for web clients
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	// Keep connection alive for mcp-remote
	w.Header().Set("Connection", "keep-alive")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST requests for MCP
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if h.debug {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		}
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if h.debug {
		fmt.Fprintf(os.Stderr, "Received HTTP MCP request: %s\n", string(body))
	}

	// Parse MCP message
	var mcpRequest mcp.Message
	if err := json.Unmarshal(body, &mcpRequest); err != nil {
		if h.debug {
			fmt.Fprintf(os.Stderr, "Error parsing MCP message: %v\n", err)
		}
		// Return JSON-RPC parse error
		errorResponse := &mcp.Message{
			JSONRPC: "2.0",
			Error: &mcp.Error{
				Code:    mcp.ParseError,
				Message: "Invalid JSON-RPC message",
			},
		}
		responseData, _ := json.Marshal(errorResponse)
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
		return
	}

	// Handle the MCP message
	mcpResponse, err := h.mcpServer.HandleMessage(r.Context(), &mcpRequest)
	if err != nil {
		if h.debug {
			fmt.Fprintf(os.Stderr, "Error handling MCP message: %v\n", err)
		}
		
		// Return JSON-RPC error response
		errorResponse := &mcp.Message{
			JSONRPC: "2.0",
			ID:      mcpRequest.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: err.Error(),
			},
		}
		
		responseData, _ := json.Marshal(errorResponse)
		w.WriteHeader(http.StatusOK) // JSON-RPC errors still return 200
		w.Write(responseData)
		return
	}

	// Return successful response
	if mcpResponse != nil {
		responseData, err := json.Marshal(mcpResponse)
		if err != nil {
			if h.debug {
				fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if h.debug {
			fmt.Fprintf(os.Stderr, "Sending HTTP MCP response: %s\n", string(responseData))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	} else {
		// No response for notifications
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

// handleHealth handles health check requests
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	healthResponse := map[string]interface{}{
		"status":    "healthy",
		"service":   "teeny-orb-mcp-server",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now()).String(),
	}
	
	json.NewEncoder(w).Encode(healthResponse)
}

// handleStatus handles status requests with detailed information
func (h *HTTPHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	statusResponse := map[string]interface{}{
		"service":   "teeny-orb-mcp-server",
		"version":   "0.1.0",
		"protocol":  "MCP 2024-11-05",
		"transport": "HTTP",
		"endpoints": map[string]string{
			"mcp":    "/mcp",
			"health": "/health",
			"status": "/status",
		},
		"capabilities": []string{
			"tools",
			"filesystem",
			"command_execution",
			"security_validation",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	json.NewEncoder(w).Encode(statusResponse)
}

// HTTPClient provides a client for making HTTP requests to MCP server
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	debug      bool
}

// NewHTTPClient creates a new HTTP client for MCP
func NewHTTPClient(baseURL string, debug bool) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: debug,
	}
}

// SendMessage sends an MCP message via HTTP
func (c *HTTPClient) SendMessage(ctx context.Context, message *mcp.Message) (*mcp.Message, error) {
	// Marshal the message
	requestData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Sending HTTP request: %s\n", string(requestData))
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "Received HTTP response: %s\n", string(responseData))
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(responseData))
	}

	// Parse response
	var mcpResponse mcp.Message
	if err := json.Unmarshal(responseData, &mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &mcpResponse, nil
}

// GetHealth checks the health of the MCP server
func (c *HTTPClient) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send health request: %w", err)
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
	}

	return health, nil
}

// GetStatus gets detailed status information from the MCP server
func (c *HTTPClient) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send status request: %w", err)
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return status, nil
}