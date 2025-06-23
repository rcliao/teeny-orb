package mcp

import (
	"context"
	"encoding/json"
)

// MCPVersion represents the MCP protocol version
const MCPVersion = "2024-11-05"

// Supported MCP protocol versions for compatibility
var SupportedMCPVersions = []string{
	"2024-11-05",
	"2025-03-26",
}

// Message represents a generic MCP message
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard MCP error codes
const (
	ParseError           = -32700
	InvalidRequest       = -32600
	MethodNotFound       = -32601
	InvalidParams        = -32602
	InternalError        = -32603
	ServerNotInitialized = -32002
	UnknownError         = -32001
)

// InitializeRequest represents the initialize request
type InitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// InitializeResponse represents the initialize response
type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     *SamplingCapability     `json:"sampling,omitempty"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
}

// SamplingCapability represents sampling capability
type SamplingCapability struct{}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema represents tool input schema
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// CallToolRequest represents a tool call request
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResponse represents a tool call response
type CallToolResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents content in MCP responses
type Content struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// ListToolsRequest represents a list tools request
type ListToolsRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// ListToolsResponse represents a list tools response
type ListToolsResponse struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string      `json:"uri"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MimeType    string      `json:"mimeType,omitempty"`
	Meta        interface{} `json:"meta,omitempty"`
}

// ReadResourceRequest represents a read resource request
type ReadResourceRequest struct {
	URI string `json:"uri"`
}

// ReadResourceResponse represents a read resource response
type ReadResourceResponse struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// MCPToolHandler defines the interface for handling MCP tool calls
type MCPToolHandler interface {
	Name() string
	Description() string
	InputSchema() InputSchema
	Handle(ctx context.Context, arguments map[string]interface{}) (*CallToolResponse, error)
}

// MCPServer defines the interface for MCP servers
type MCPServer interface {
	// Initialize initializes the server
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	
	// RegisterTool registers a tool handler
	RegisterTool(handler MCPToolHandler) error
	
	// ListTools lists available tools
	ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error)
	
	// CallTool calls a tool
	CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error)
	
	// Serve starts the server
	Serve(ctx context.Context) error
	
	// Close closes the server
	Close() error
}

// MCPClient defines the interface for MCP clients
type MCPClient interface {
	// Connect connects to the server
	Connect(ctx context.Context) error
	
	// Initialize performs the initialization handshake
	Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
	
	// ListTools lists available tools from the server
	ListTools(ctx context.Context) (*ListToolsResponse, error)
	
	// CallTool calls a tool on the server
	CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error)
	
	// Close closes the connection
	Close() error
}

// Transport defines the interface for MCP transport layers
type Transport interface {
	// Send sends a message
	Send(ctx context.Context, msg *Message) error
	
	// Receive receives a message
	Receive(ctx context.Context) (*Message, error)
	
	// Close closes the transport
	Close() error
}