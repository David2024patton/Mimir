package llm

import (
	"sort"
	"sync"
)

// Usage is the token accounting for one completion (F37.1).
type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

// Total returns prompt + completion tokens.
func (u Usage) Total() int { return u.PromptTokens + u.CompletionTokens }

// IsZero reports whether no usage was recorded.
func (u Usage) IsZero() bool { return u.PromptTokens == 0 && u.CompletionTokens == 0 }

// ModelUsage is the accumulated usage for one model.
type ModelUsage struct {
	Model            string `json:"model"`
	Local            bool   `json:"local"`
	Requests         int    `json:"requests"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

// Tracker accumulates token usage per model. It is safe for concurrent use (the
// streaming server handles many turns at once). It feeds the per-project dashboard
// (E69 / F52): total + per-model tokens, sorted most-to-least, local vs cloud.
type Tracker struct {
	mu      sync.Mutex
	byModel map[string]*ModelUsage
}

// NewTracker returns an empty usage tracker.
func NewTracker() *Tracker { return &Tracker{byModel: map[string]*ModelUsage{}} }

// Record adds one completion's usage to the model's totals. Zero usage is ignored.
func (t *Tracker) Record(model string, local bool, u Usage) {
	if u.IsZero() {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	mu, ok := t.byModel[model]
	if !ok {
		mu = &ModelUsage{Model: model, Local: local}
		t.byModel[model] = mu
	}
	mu.Requests++
	mu.PromptTokens += u.PromptTokens
	mu.CompletionTokens += u.CompletionTokens
	mu.TotalTokens += u.Total()
}

// Snapshot returns per-model usage sorted most-tokens to least (F52.3).
func (t *Tracker) Snapshot() []ModelUsage {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]ModelUsage, 0, len(t.byModel))
	for _, mu := range t.byModel {
		out = append(out, *mu)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TotalTokens > out[j].TotalTokens })
	return out
}

// TotalTokens returns the sum of all tokens across every model.
func (t *Tracker) TotalTokens() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	total := 0
	for _, mu := range t.byModel {
		total += mu.TotalTokens
	}
	return total
}
