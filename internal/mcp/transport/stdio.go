package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rcliao/teeny-orb/internal/mcp"
)

// StdioTransport implements MCP transport over stdin/stdout
type StdioTransport struct {
	stdin  io.Reader
	stdout io.Writer
	scanner *bufio.Scanner
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// NewStdioTransportWithStreams creates a stdio transport with custom streams
func NewStdioTransportWithStreams(stdin io.Reader, stdout io.Writer) *StdioTransport {
	return &StdioTransport{
		stdin:   stdin,
		stdout:  stdout,
		scanner: bufio.NewScanner(stdin),
	}
}

// Send sends a message over stdout
func (s *StdioTransport) Send(ctx context.Context, msg *mcp.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Write JSON-RPC message followed by newline
	_, err = fmt.Fprintf(s.stdout, "%s\n", data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	
	return nil
}

// Receive receives a message from stdin
func (s *StdioTransport) Receive(ctx context.Context) (*mcp.Message, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	// Read line from stdin
	if !s.scanner.Scan() {
		if err := s.scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanner error: %w", err)
		}
		return nil, io.EOF
	}
	
	line := s.scanner.Text()
	if line == "" {
		return s.Receive(ctx) // Skip empty lines
	}
	
	// Parse JSON-RPC message
	var msg mcp.Message
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	return &msg, nil
}

// Close closes the transport
func (s *StdioTransport) Close() error {
	// For stdio transport, we don't close stdin/stdout
	// as they might be used by other parts of the application
	return nil
}