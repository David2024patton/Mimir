package tools

import "context"

// Tool is one capability the agent can call.
type Tool interface {
	Name() string
	Description() string
	Parameters() any // JSON Schema
	Run(ctx context.Context, args map[string]any) (string, error)
}

// Registry holds the available tools (native + MCP).
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) All() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}
