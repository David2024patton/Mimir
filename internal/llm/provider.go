package llm

import "context"

// Message is a chat message.
type Message struct {
	Role      string     `json:"role"` // system | user | assistant | tool
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a model-issued tool invocation.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// GenerateRequest is a provider-agnostic completion request.
type GenerateRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolSchema
	Stream   bool
}

// ToolSchema describes a tool to the model (JSON Schema).
type ToolSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// StreamEvent is one chunk of a streamed response.
type StreamEvent struct {
	Type     string // token | tool_call | usage | done | error
	Delta    string
	ToolCall *ToolCall
	Usage    *Usage
	Err      error
}

// Usage reports token counts.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// GenerateResponse is a non-streamed completion.
type GenerateResponse struct {
	Message Message
	Usage   Usage
}

// HTTPRequest is a dialect-encoded outbound request.
type HTTPRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
}

// Dialect encodes/decodes one API wire format (OpenAI-compatible, Anthropic, ...).
type Dialect interface {
	Name() string
	EncodeRequest(req GenerateRequest, apiKey string) (*HTTPRequest, error)
	DecodeStream(chunk []byte) ([]StreamEvent, error)
}

// Provider talks to one model backend via a Dialect.
type Provider interface {
	ID() string
	Dialect() Dialect
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	Stream(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error)
}
