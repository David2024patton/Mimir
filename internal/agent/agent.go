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
	Usage    *llm.Tracker // optional: records token usage for the dashboard (E69)
	Local    bool         // whether the provider is local (local tokens cost $0)
}

// Step records one tool execution in a turn (used by RunTrace / the trace CLI).
type Step struct {
	Name   string `json:"name"`
	Args   string `json:"args"`
	Result string `json:"result"`
	Err    string `json:"err,omitempty"`
}

// Result is the full outcome of a turn: the final reply, the tool trace, and the
// memories that were recalled into the prompt (so memory is observable/testable).
type Result struct {
	Reply    string
	Trace    []Step
	Recalled []string
}

// Event kinds emitted by RunStream (F2.1 streaming).
const (
	EventMemory = "memory" // recalled memory (Text = content)
	EventToken  = "token"  // a delta of the model's reply (Text = delta)
	EventTool   = "tool"   // a tool execution (Step)
	EventDone   = "done"   // turn complete (Reply = final answer)
	EventError  = "error"  // turn failed (Err)
)

// Event is one streaming update from a turn, surfaced to live UIs by RunStream.
type Event struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Step  *Step  `json:"step,omitempty"`
	Reply string `json:"reply,omitempty"`
	Err   string `json:"err,omitempty"`
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
		if a.cfg.Usage != nil {
			a.cfg.Usage.Record(a.cfg.Model, a.cfg.Local, resp.Usage)
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

// Run executes one conversational turn and returns the final text reply.
func (a *Agent) Run(ctx context.Context, input string) (string, error) {
	r, err := a.runTurn(ctx, input)
	return r.Reply, err
}

// RunTrace is like Run but also returns the trace of tool executions.
func (a *Agent) RunTrace(ctx context.Context, input string) (string, []Step, error) {
	r, err := a.runTurn(ctx, input)
	return r.Reply, r.Trace, err
}

// RunFull returns the complete Result, including the memories recalled into the
// prompt - used by the trace CLI and the cross-run memory tests.
func (a *Agent) RunFull(ctx context.Context, input string) (Result, error) {
	return a.runTurn(ctx, input)
}

// RunStream runs one conversational turn while streaming updates to emit as they
// happen: recalled memories, token deltas of the model's reply, tool executions, and
// a final done event carrying the full reply (F2.1 streaming + tool calling). It
// mirrors RunFull's recall -> infer -> act -> remember cycle, but the answer streams
// token by token instead of arriving all at once. The doom-loop guard caps steps.
func (a *Agent) RunStream(ctx context.Context, input string, emit func(Event)) error {
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		emit(Event{Type: EventError, Err: err.Error()})
		return err
	}
	for _, m := range memories {
		emit(Event{Type: EventMemory, Text: m.Content})
	}
	msgs := []llm.Message{
		{Role: "system", Content: systemPrompt(memories, a.schemas())},
		{Role: "user", Content: input},
	}
	schemas := a.schemas()
	var finalReply string
	for step := 0; step < maxToolSteps; step++ {
		ch, err := a.cfg.Provider.Stream(ctx, llm.GenerateRequest{
			Model: a.cfg.Model, Messages: msgs, Tools: schemas,
		})
		if err != nil {
			emit(Event{Type: EventError, Err: err.Error()})
			return err
		}
		var content strings.Builder
		var toolCalls []llm.ToolCall
		var usage llm.Usage
		for ev := range ch {
			if ev.Err != nil {
				emit(Event{Type: EventError, Err: ev.Err.Error()})
				return ev.Err
			}
			if ev.Delta != "" {
				content.WriteString(ev.Delta)
				emit(Event{Type: EventToken, Text: ev.Delta})
			}
			if ev.Done {
				toolCalls = ev.ToolCalls
				usage = ev.Usage
			}
		}
		if a.cfg.Usage != nil {
			a.cfg.Usage.Record(a.cfg.Model, a.cfg.Local, usage)
		}
		text := content.String()
		finalReply = text
		msgs = append(msgs, llm.Message{Role: "assistant", Content: text, ToolCalls: toolCalls})
		if len(toolCalls) == 0 {
			emit(Event{Type: EventDone, Reply: finalReply})
			return nil
		}
		for _, tc := range toolCalls {
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
			emit(Event{Type: EventTool, Step: &st})
			msgs = append(msgs, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: st.Result})
		}
	}
	emit(Event{Type: EventDone, Reply: finalReply})
	return nil
}

func (a *Agent) runTurn(ctx context.Context, input string) (Result, error) {
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		return Result{}, err
	}
	recalled := make([]string, len(memories))
	for i, m := range memories {
		recalled[i] = m.Content
	}
	msgs := []llm.Message{
		{Role: "system", Content: systemPrompt(memories, a.schemas())},
		{Role: "user", Content: input},
	}
	reply, trace, err := a.runLoop(ctx, msgs)
	if err != nil {
		return Result{Trace: trace, Recalled: recalled}, err
	}
	return Result{Reply: reply, Trace: trace, Recalled: recalled}, nil
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
