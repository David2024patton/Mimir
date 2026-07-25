// Package agent is the agent loop (F2): it drives a model provider and a tool
// registry through the gather -> infer -> act -> verify cycle, consulting the
// Cortex for memory.
package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// maxToolSteps caps the infer->act loop so a misbehaving model can't spin forever
// (doom-loop protection, F4.3).
const maxToolSteps = 8

// Config wires an agent's dependencies.
type Config struct {
	Provider llm.Provider
	Tools    *tools.Registry
	Cortex   cortex.Store
	Model    string
}

// Step records one tool execution in a turn (used by RunTrace / the trace CLI).
type Step struct {
	Name   string
	Args   string
	Result string
	Err    string
}

// Agent runs the gather -> infer -> act -> verify loop.
type Agent struct {
	cfg Config
}

// New returns an Agent wired to the given dependencies.
func New(cfg Config) *Agent {
	return &Agent{cfg: cfg}
}

// schemas converts the registered tools into model-facing tool schemas.
func (a *Agent) schemas() []llm.ToolSchema {
	if a.cfg.Tools == nil {
		return nil
	}
	var out []llm.ToolSchema
	for _, t := range a.cfg.Tools.All() {
		out = append(out, llm.ToolSchema{
			Name: t.Name(), Description: t.Description(), Parameters: t.Parameters(),
		})
	}
	return out
}

// runLoop drives the infer -> act cycle until the model returns text or the step
// cap is hit. It returns the final text reply and a trace of tool executions.
func (a *Agent) runLoop(ctx context.Context, msgs []llm.Message) (string, []Step, error) {
	schemas := a.schemas()
	var trace []Step
	var last string
	for step := 0; step < maxToolSteps; step++ {
		resp, err := a.cfg.Provider.Generate(ctx, llm.GenerateRequest{
			Model: a.cfg.Model, Messages: msgs, Tools: schemas,
		})
		if err != nil {
			return "", trace, err
		}
		last = resp.Content
		msgs = append(msgs, llm.Message{Role: "assistant", Content: resp.Content, ToolCalls: resp.ToolCalls})
		if len(resp.ToolCalls) == 0 {
			return resp.Content, trace, nil
		}
		for _, tc := range resp.ToolCalls {
			st := Step{Name: tc.Name, Args: tc.Arguments}
			tool, ok := a.cfg.Tools.Get(tc.Name)
			if !ok {
				st.Err = "unknown tool: " + tc.Name
				st.Result = st.Err
			} else {
				var args map[string]any
				if tc.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &args)
				}
				r, err := tool.Run(ctx, args)
				st.Result = r
				if err != nil {
					st.Err = err.Error()
				}
			}
			trace = append(trace, st)
			msgs = append(msgs, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: st.Result})
		}
	}
	return last, trace, nil
}

// Run executes one conversational turn: recall memory, run the tool loop, remember
// the exchange, and return the final text reply.
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	reply, _, err := a.runTurn(ctx, input)
	return reply, err
}

// RunTrace is like Run but also returns the trace of tool executions.
func (a *Agent) RunTrace(ctx context.Context, input string) (string, []Step, error) {
	return a.runTurn(ctx, input)
}

func (a *Agent) runTurn(ctx context.Context, input string) (string, []Step, error) {
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		return "", nil, err
	}
	msgs := []llm.Message{
		{Role: "system", Content: systemPrompt(memories, a.schemas())},
		{Role: "user", Content: input},
	}
	reply, trace, err := a.runLoop(ctx, msgs)
	if err != nil {
		return "", trace, err
	}
	_, _ = a.cfg.Cortex.PutNeuron(ctx, cortex.Neuron{
		Kind: cortex.KindMemory, Layer: "experience", Title: "exchange",
		Content: input + "\n" + reply, Decay: 1.0,
	})
	return reply, trace, nil
}

// systemPrompt assembles the system prompt with recalled memory + a tool hint.
func systemPrompt(memories []cortex.Neuron, schemas []llm.ToolSchema) string {
	p := "You are Mímir, an agent that remembers. Use what you recall."
	if len(memories) > 0 {
		p += "\n\nRelevant memory:"
		for _, m := range memories {
			p += "\n- " + m.Content
		}
	}
	if len(schemas) > 0 {
		names := make([]string, len(schemas))
		for i, s := range schemas {
			names[i] = s.Name
		}
		p += "\n\nYou have these tools - call them when useful: " + strings.Join(names, ", ") + "."
	}
	return p
}
