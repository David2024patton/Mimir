// Package llm defines the model provider abstraction. Mímir talks to any model
// (cloud or local) through this interface - BYOK or Ollama (F1).
package llm

import "context"

// Message is one turn in a conversation.
type Message struct {
	Role       string     // system | user | assistant | tool
	Content    string
	ToolCalls  []ToolCall // assistant messages may carry tool calls
	ToolCallID string     // tool-role messages reference the call id
}

// ToolCall is a model-issued tool invocation.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON object string
}

// ToolSchema describes a tool to the model.
type ToolSchema struct {
	Name        string
	Description string
	Parameters  any // a JSON Schema object
}

// GenerateRequest is a provider-agnostic completion request.
type GenerateRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolSchema
	Stream   bool
}

// GenerateResponse is a provider-agnostic completion response.
type GenerateResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        Usage
}

// StreamEvent is one chunk of a streamed response. Text arrives as Delta events;
// the final event has Done set and carries any ToolCalls the model issued
// (accumulated across the stream's fragmented tool-call deltas) and the token Usage.
type StreamEvent struct {
	Delta     string
	ToolCalls []ToolCall
	Usage     Usage
	Done      bool
	Err       error
}

// Provider talks to one model backend. Implementations wrap OpenAI-compatible
// APIs (cloud) or Ollama (local).
type Provider interface {
	// ID returns the provider id (e.g. "ollama", "openai").
	ID() string
	// Generate returns a full (non-streamed) completion, including any tool calls.
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	// Stream returns a channel of streamed chunks.
	Stream(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error)
}

// Registry maps provider ids to Providers.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: map[string]Provider{}}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.providers[p.ID()] = p
}

// Get returns a provider by id.
func (r *Registry) Get(id string) (Provider, bool) {
	p, ok := r.providers[id]
	return p, ok
}
