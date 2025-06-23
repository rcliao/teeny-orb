package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rcliao/teeny-orb/internal/mcp"
)

// Server implements the MCP server interface
type Server struct {
	info         mcp.ServerInfo
	capabilities mcp.ServerCapabilities
	tools        map[string]mcp.MCPToolHandler
	initialized  bool
	mutex        sync.RWMutex
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		info: mcp.ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsCapability{
				ListChanged: false,
			},
			Logging: &mcp.LoggingCapability{},
		},
		tools: make(map[string]mcp.MCPToolHandler),
	}
}

// Initialize handles the initialization request
func (s *Server) Initialize(ctx context.Context, req *mcp.InitializeRequest) (*mcp.InitializeResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Accept any reasonable protocol version for maximum compatibility
	// Log for debugging but don't reject
	if req.ProtocolVersion != "" {
		fmt.Printf("DEBUG: Client requested protocol version: %s\n", req.ProtocolVersion)
	}

	s.initialized = true

	// Respond with the client's requested version if supported, otherwise use our default
	responseVersion := req.ProtocolVersion
	if responseVersion == "" {
		responseVersion = mcp.MCPVersion
	}

	return &mcp.InitializeResponse{
		ProtocolVersion: responseVersion,
		Capabilities:    s.capabilities,
		ServerInfo:      s.info,
	}, nil
}

// RegisterTool registers a tool handler
func (s *Server) RegisterTool(handler mcp.MCPToolHandler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	name := handler.Name()
	if _, exists := s.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}

	s.tools[name] = handler
	return nil
}

// ListTools lists all available tools
func (s *Server) ListTools(ctx context.Context, req *mcp.ListToolsRequest) (*mcp.ListToolsResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, fmt.Errorf("server not initialized")
	}

	tools := make([]mcp.Tool, 0, len(s.tools))
	for _, handler := range s.tools {
		tools = append(tools, mcp.Tool{
			Name:        handler.Name(),
			Description: handler.Description(),
			InputSchema: handler.InputSchema(),
		})
	}

	return &mcp.ListToolsResponse{
		Tools: tools,
	}, nil
}

// CallTool executes a tool call
func (s *Server) CallTool(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResponse, error) {
	s.mutex.RLock()
	handler, exists := s.tools[req.Name]
	s.mutex.RUnlock()

	if !exists {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool not found: %s", req.Name),
				},
			},
			IsError: true,
		}, nil
	}

	if !s.initialized {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Server not initialized",
				},
			},
			IsError: true,
		}, nil
	}

	return handler.Handle(ctx, req.Arguments)
}

// HandleMessage processes incoming MCP messages
func (s *Server) HandleMessage(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(ctx, msg)
	case "tools/list":
		return s.handleListTools(ctx, msg)
	case "tools/call":
		return s.handleCallTool(ctx, msg)
	default:
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.MethodNotFound,
				Message: fmt.Sprintf("Method not found: %s", msg.Method),
			},
		}, nil
	}
}

func (s *Server) handleInitialize(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
	var req mcp.InitializeRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InvalidParams,
				Message: "Invalid initialize parameters",
			},
		}, nil
	}

	resp, err := s.Initialize(ctx, &req)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: err.Error(),
			},
		}, nil
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: "Failed to marshal response",
			},
		}, nil
	}

	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleListTools(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
	var req mcp.ListToolsRequest
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &req); err != nil {
			return &mcp.Message{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Error: &mcp.Error{
					Code:    mcp.InvalidParams,
					Message: "Invalid list tools parameters",
				},
			}, nil
		}
	}

	resp, err := s.ListTools(ctx, &req)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: err.Error(),
			},
		}, nil
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: "Failed to marshal response",
			},
		}, nil
	}

	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

func (s *Server) handleCallTool(ctx context.Context, msg *mcp.Message) (*mcp.Message, error) {
	var req mcp.CallToolRequest
	if err := json.Unmarshal(msg.Params, &req); err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InvalidParams,
				Message: "Invalid call tool parameters",
			},
		}, nil
	}

	resp, err := s.CallTool(ctx, &req)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: err.Error(),
			},
		}, nil
	}

	result, err := json.Marshal(resp)
	if err != nil {
		return &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Error: &mcp.Error{
				Code:    mcp.InternalError,
				Message: "Failed to marshal response",
			},
		}, nil
	}

	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}, nil
}

// Serve starts the server (stub implementation)
func (s *Server) Serve(ctx context.Context) error {
	// In a real implementation, this would start the transport layer
	// and handle incoming connections
	<-ctx.Done()
	return ctx.Err()
}

// Close closes the server
func (s *Server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.initialized = false
	s.tools = make(map[string]mcp.MCPToolHandler)
	return nil
}