// Package tools defines the built-in system tools (F3): terminal, filesystem,
// code execution, and the tool registry the agent calls into.
package tools

import "context"

// Tool is one capability the agent can call.
type Tool interface {
	// Name returns the tool name (e.g. "bash", "read_file").
	Name() string
	// Description explains the tool to the model.
	Description() string
	// Parameters returns the tool's JSON Schema (an object schema).
	Parameters() any
	// Run executes the tool with the given arguments and returns its output.
	Run(ctx context.Context, args map[string]any) (string, error)
}

// Registry holds the available tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry returns an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns every registered tool.
func (r *Registry) All() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// IsWriteTool reports whether a tool mutates files or executes shell commands.
// In read-only ("ro") mode, write tools are hidden from the model's tool schema
// so smaller models cannot hallucinate edits (smolcoder pattern).
func IsWriteTool(name string) bool {
	switch name {
	case "write_file", "edit_block", "bash":
		return true
	default:
		return false
	}
}

// ForMode returns the tools visible for the given permission mode ("ro", "edit", "bypass").
func (r *Registry) ForMode(mode string) []Tool {
	if mode != "ro" {
		return r.All()
	}
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		if !IsWriteTool(t.Name()) {
			out = append(out, t)
		}
	}
	return out
}
