// Package llm defines the model provider abstraction. Mímir talks to any model
// (cloud or local) through this interface - BYOK or Ollama (F1).
package llm

import "context"

// Message is one turn in a conversation.
type Message struct {
	Role    string // system | user | assistant | tool
	Content string
}

// GenerateRequest is a provider-agnostic completion request.
type GenerateRequest struct {
	Model    string
	Messages []Message
	Stream   bool
}

// StreamEvent is one chunk of a streamed response.
type StreamEvent struct {
	Delta string
	Done  bool
	Err   error
}

// Provider talks to one model backend. Implementations wrap OpenAI-compatible
// APIs (cloud) or Ollama (local).
type Provider interface {
	// ID returns the provider id (e.g. "ollama", "openai").
	ID() string
	// Generate returns a full (non-streamed) completion.
	Generate(ctx context.Context, req GenerateRequest) (string, error)
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
