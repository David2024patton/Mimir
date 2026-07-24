package llm

import (
	"fmt"
	"strings"
)

// Registry maps provider IDs to Providers and routes by model string.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: map[string]Provider{}}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.ID()] = p
}

func (r *Registry) Get(id string) (Provider, bool) {
	p, ok := r.providers[id]
	return p, ok
}

// Route picks a provider for a model string like "openai/gpt-..." or "anthropic/claude-...".
func (r *Registry) Route(model string) (Provider, error) {
	if p, ok := r.providers[model]; ok {
		return p, nil
	}
	// TODO(E2.1): parse "provider/model" prefix and select the matching provider.
	prefix := model
	if i := strings.Index(model, "/"); i >= 0 {
		prefix = model[:i]
	}
	if p, ok := r.providers[prefix]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("no provider for model %q", model)
}
