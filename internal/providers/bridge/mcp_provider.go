package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/rcliao/teeny-orb/internal/mcp"
	"github.com/rcliao/teeny-orb/internal/mcp/server"
	"github.com/rcliao/teeny-orb/internal/mcp/tools"
	"github.com/rcliao/teeny-orb/internal/providers"
)

// MCPToolProvider bridges MCP server to the ToolProvider interface
type MCPToolProvider struct {
	server     *server.Server
	initialized bool
	mutex      sync.RWMutex
}

// NewMCPToolProvider creates a new MCP tool provider bridge
func NewMCPToolProvider() *MCPToolProvider {
	mcpServer := server.NewServer("teeny-orb-experiment", "0.1.0")
	
	// Register default tools
	fsTools := tools.NewFileSystemTool("/workspace")
	cmdTool := tools.NewCommandTool([]string{"ls", "pwd", "echo", "cat"})
	
	mcpServer.RegisterTool(fsTools)
	mcpServer.RegisterTool(cmdTool)
	
	return &MCPToolProvider{
		server: mcpServer,
	}
}

// initialize performs MCP initialization if not already done
func (m *MCPToolProvider) initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.initialized {
		return nil
	}
	
	// Simulate MCP initialization
	initReq := &mcp.InitializeRequest{
		ProtocolVersion: mcp.MCPVersion,
		Capabilities: mcp.ClientCapabilities{
			Experimental: make(map[string]interface{}),
		},
		ClientInfo: mcp.ClientInfo{
			Name:    "teeny-orb-experiment",
			Version: "0.1.0",
		},
	}
	
	_, err := m.server.Initialize(context.Background(), initReq)
	if err != nil {
		return fmt.Errorf("MCP initialization failed: %w", err)
	}
	
	m.initialized = true
	return nil
}

// RegisterTool registers a tool by bridging it to MCP
func (m *MCPToolProvider) RegisterTool(tool providers.Tool) error {
	if err := m.initialize(); err != nil {
		return err
	}
	
	// Create MCP tool wrapper
	mcpTool := &toolBridge{tool: tool}
	return m.server.RegisterTool(mcpTool)
}

// ListTools returns all registered tools
func (m *MCPToolProvider) ListTools() []providers.Tool {
	if err := m.initialize(); err != nil {
		return []providers.Tool{}
	}
	
	// Get tools from MCP server
	listReq := &mcp.ListToolsRequest{}
	listResp, err := m.server.ListTools(context.Background(), listReq)
	if err != nil {
		return []providers.Tool{}
	}
	
	// Convert MCP tools back to provider tools (simplified)
	tools := make([]providers.Tool, len(listResp.Tools))
	for i, mcpTool := range listResp.Tools {
		tools[i] = &mcpToolWrapper{
			name:        mcpTool.Name,
			description: mcpTool.Description,
			server:      m.server,
		}
	}
	
	return tools
}

// CallTool executes a tool through MCP protocol
func (m *MCPToolProvider) CallTool(ctx context.Context, name string, args map[string]interface{}) (*providers.ToolResult, error) {
	if err := m.initialize(); err != nil {
		return &providers.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("MCP initialization failed: %v", err),
		}, nil
	}
	
	// Convert to MCP call tool request
	callReq := &mcp.CallToolRequest{
		Name:      name,
		Arguments: args,
	}
	
	// Execute through MCP protocol (with serialization overhead)
	callResp, err := m.server.CallTool(ctx, callReq)
	if err != nil {
		return &providers.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	// Convert MCP response back to ToolResult
	if callResp.IsError {
		errorMsg := "Unknown MCP error"
		if len(callResp.Content) > 0 {
			errorMsg = callResp.Content[0].Text
		}
		return &providers.ToolResult{
			Success: false,
			Error:   errorMsg,
		}, nil
	}
	
	// Aggregate output from content
	output := ""
	data := make(map[string]interface{})
	for _, content := range callResp.Content {
		if content.Type == "text" {
			output += content.Text
		}
		if content.Data != nil {
			data["mcp_data"] = content.Data
		}
	}
	
	return &providers.ToolResult{
		Success: true,
		Data:    data,
		Output:  output,
	}, nil
}

// Close closes the MCP server
func (m *MCPToolProvider) Close() error {
	return m.server.Close()
}

// toolBridge bridges a providers.Tool to mcp.MCPToolHandler
type toolBridge struct {
	tool providers.Tool
}

func (t *toolBridge) Name() string {
	return t.tool.Name()
}

func (t *toolBridge) Description() string {
	return t.tool.Description()
}

func (t *toolBridge) InputSchema() mcp.InputSchema {
	// Create basic schema based on tool name
	// In a real implementation, this would be more sophisticated
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"args": map[string]interface{}{
				"type":        "object",
				"description": "Tool arguments",
			},
		},
	}
}

func (t *toolBridge) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	result, err := t.tool.Execute(ctx, arguments)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool execution error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}
	
	if !result.Success {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: result.Error,
				},
			},
			IsError: true,
		}, nil
	}
	
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: result.Output,
				Data: result.Data,
			},
		},
		IsError: false,
	}, nil
}

// mcpToolWrapper wraps an MCP tool as a providers.Tool
type mcpToolWrapper struct {
	name        string
	description string
	server      *server.Server
}

func (m *mcpToolWrapper) Name() string {
	return m.name
}

func (m *mcpToolWrapper) Description() string {
	return m.description
}

func (m *mcpToolWrapper) Execute(ctx context.Context, args map[string]interface{}) (*providers.ToolResult, error) {
	callReq := &mcp.CallToolRequest{
		Name:      m.name,
		Arguments: args,
	}
	
	callResp, err := m.server.CallTool(ctx, callReq)
	if err != nil {
		return &providers.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	
	if callResp.IsError {
		errorMsg := "Unknown error"
		if len(callResp.Content) > 0 {
			errorMsg = callResp.Content[0].Text
		}
		return &providers.ToolResult{
			Success: false,
			Error:   errorMsg,
		}, nil
	}
	
	output := ""
	for _, content := range callResp.Content {
		if content.Type == "text" {
			output += content.Text
		}
	}
	
	return &providers.ToolResult{
		Success: true,
		Output:  output,
	}, nil
}