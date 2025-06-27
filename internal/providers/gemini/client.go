package gemini

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rcliao/teeny-orb/internal/providers"
)

// GeminiClient implements the AIProvider interface for Google Gemini
type GeminiClient struct {
	apiKey      string
	baseURL     string
	model       string
	httpClient  *http.Client
	toolProvider providers.ToolProvider
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToolProvider sets the tool provider for function calling
func (g *GeminiClient) SetToolProvider(provider providers.ToolProvider) {
	g.toolProvider = provider
}

// Chat sends a chat request to Gemini
func (g *GeminiClient) Chat(ctx context.Context, request *providers.ChatRequest) (*providers.ChatResponse, error) {
	// Convert provider request to Gemini format
	geminiRequest := g.convertToGeminiRequest(request)
	
	// Make API call
	respData, err := g.makeAPICall(ctx, geminiRequest)
	if err != nil {
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}
	
	// Convert response back to provider format
	return g.convertFromGeminiResponse(respData, request.Model)
}

// ChatStream implements streaming (simplified for testing)
func (g *GeminiClient) ChatStream(ctx context.Context, request *providers.ChatRequest) (<-chan *providers.StreamChunk, error) {
	ch := make(chan *providers.StreamChunk, 3)
	
	go func() {
		defer close(ch)
		
		// For testing, simulate streaming by calling regular chat and chunking the response
		response, err := g.Chat(ctx, request)
		if err != nil {
			ch <- &providers.StreamChunk{
				Error: err,
				Done:  true,
			}
			return
		}
		
		// Split content into chunks
		content := response.Content
		chunkSize := 50
		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}
			
			ch <- &providers.StreamChunk{
				Content: content[i:end],
				Done:    end == len(content),
			}
			
			// Small delay to simulate streaming
			time.Sleep(10 * time.Millisecond)
		}
	}()
	
	return ch, nil
}

// CountTokens estimates token count (simplified implementation)
func (g *GeminiClient) CountTokens(text string) (int, error) {
	// Rough estimation: ~1.3 tokens per word for English
	wordCount := len(text) / 5 // Approximate words
	return int(float64(wordCount) * 1.3), nil
}

// GetModel returns model information
func (g *GeminiClient) GetModel() *providers.ModelInfo {
	return &providers.ModelInfo{
		Name:         g.model,
		Provider:     "gemini",
		MaxTokens:    1000000, // Gemini 1.5 Pro has large context
		SupportsTools: true,
	}
}

// makeAPICall performs the actual HTTP request to Gemini API
func (g *GeminiClient) makeAPICall(ctx context.Context, request *GeminiRequest) (*GeminiResponse, error) {
	// For testing purposes, return simulated response instead of real API call
	// In production, this would make actual HTTP requests
	
	return &GeminiResponse{
		Candidates: []Candidate{
			{
				Content: Content{
					Parts: []Part{
						{
							Text: "This is a simulated Gemini response for cross-provider testing. " +
								 "In a real implementation, this would call the actual Gemini API.",
						},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: UsageMetadata{
			PromptTokenCount:     estimateTokens(request),
			CandidatesTokenCount: 50,
			TotalTokenCount:      estimateTokens(request) + 50,
		},
	}, nil
}

// convertToGeminiRequest converts provider request to Gemini API format
func (g *GeminiClient) convertToGeminiRequest(request *providers.ChatRequest) *GeminiRequest {
	contents := make([]Content, len(request.Messages))
	
	for i, msg := range request.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		
		contents[i] = Content{
			Role: role,
			Parts: []Part{
				{Text: msg.Content},
			},
		}
	}
	
	geminiRequest := &GeminiRequest{
		Contents: contents,
		GenerationConfig: GenerationConfig{
			Temperature:     0.7,
			TopK:           40,
			TopP:           0.95,
			MaxOutputTokens: 2048,
		},
	}
	
	// Add tools if available
	if g.toolProvider != nil && len(request.Tools) > 0 {
		geminiRequest.Tools = g.convertTools(request.Tools)
	}
	
	return geminiRequest
}

// convertFromGeminiResponse converts Gemini response to provider format
func (g *GeminiClient) convertFromGeminiResponse(response *GeminiResponse, model string) (*providers.ChatResponse, error) {
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}
	
	candidate := response.Candidates[0]
	content := ""
	var toolCalls []providers.ToolCall
	
	// Extract text content and tool calls
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			content += part.Text
		}
		if part.FunctionCall != nil {
			toolCalls = append(toolCalls, providers.ToolCall{
				ID:        fmt.Sprintf("call_%d", len(toolCalls)),
				Name:      part.FunctionCall.Name,
				Arguments: part.FunctionCall.Args,
			})
		}
	}
	
	return &providers.ChatResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Usage: providers.Usage{
			PromptTokens:     response.UsageMetadata.PromptTokenCount,
			CompletionTokens: response.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      response.UsageMetadata.TotalTokenCount,
		},
		Model: model,
	}, nil
}

// convertTools converts provider tools to Gemini function declarations
func (g *GeminiClient) convertTools(tools []providers.ToolDefinition) []Tool {
	geminiTools := make([]Tool, len(tools))
	
	for i, tool := range tools {
		geminiTools[i] = Tool{
			FunctionDeclarations: []FunctionDeclaration{
				{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			},
		}
	}
	
	return geminiTools
}

// estimateTokens provides a rough token estimate for testing
func estimateTokens(request *GeminiRequest) int {
	totalChars := 0
	for _, content := range request.Contents {
		for _, part := range content.Parts {
			totalChars += len(part.Text)
		}
	}
	// Rough estimation: ~4 characters per token
	return totalChars / 4
}

// Gemini API request/response structures

type GeminiRequest struct {
	Contents         []Content        `json:"contents"`
	Tools            []Tool           `json:"tools,omitempty"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text         string        `json:"text,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
}

type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations"`
}

type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	TopK           int     `json:"topK"`
	TopP           float64 `json:"topP"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type GeminiResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}