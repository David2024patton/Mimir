// Package agent is the agent loop (F2): it drives a model provider and a tool
// registry through the gather -> infer -> act -> verify cycle, consulting the
// Cortex for memory.
package agent

import (
	"context"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// Config wires an agent's dependencies.
type Config struct {
	Provider llm.Provider
	Tools    *tools.Registry
	Cortex   cortex.Store
	Model    string
}

// Agent runs the gather -> infer -> act -> verify loop.
type Agent struct {
	cfg Config
}

// New returns an Agent wired to the given dependencies.
func New(cfg Config) *Agent {
	return &Agent{cfg: cfg}
}

// Run executes one conversational turn: it recalls relevant memory from the
// Cortex, asks the model, runs any tool calls, and returns the reply.
//
// This is the walking-skeleton shape; the full SDLC-aware loop (F43) builds on
// it.
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	// 1. Gather: recall relevant memory from the Cortex.
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		return "", err
	}

	// 2. Build the message list (system + recalled memory + user input).
	msgs := []llm.Message{{Role: "system", Content: systemPrompt(memories)}}
	msgs = append(msgs, llm.Message{Role: "user", Content: input})

	// 3. Infer: ask the model.
	reply, err := a.cfg.Provider.Generate(ctx, llm.GenerateRequest{
		Model:    a.cfg.Model,
		Messages: msgs,
	})
	if err != nil {
		return "", err
	}

	// 4. Remember: store the exchange as a new neuron.
	_, _ = a.cfg.Cortex.PutNeuron(ctx, cortex.Neuron{
		Kind:    cortex.KindMemory,
		Layer:   "experience",
		Title:   "exchange",
		Content: input + "\n" + reply,
		Decay:   1.0,
	})

	return reply, nil
}

// systemPrompt assembles the system prompt with recalled memory.
func systemPrompt(memories []cortex.Neuron) string {
	p := "You are Mímir, an agent that remembers. Use what you recall."
	if len(memories) > 0 {
		p += "\n\nRelevant memory:"
		for _, m := range memories {
			p += "\n- " + m.Content
		}
	}
	return p
}
