package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// fakeProvider is a deterministic llm.Provider: first call returns a tool call,
// the second returns final text. It records the messages it sees each call so the
// test can prove the tool result was fed back into the loop.
type fakeProvider struct {
	calls int
	seen  [][]llm.Message
}

func (f *fakeProvider) ID() string { return "fake" }

func (f *fakeProvider) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	f.seen = append(f.seen, req.Messages)
	f.calls++
	if f.calls == 1 {
		return llm.GenerateResponse{
			ToolCalls:    []llm.ToolCall{{ID: "c1", Name: "bash", Arguments: `{"command":"echo MIMIRX"}`}},
			FinishReason: "tool_calls",
		}, nil
	}
	return llm.GenerateResponse{Content: "all done", FinishReason: "stop"}, nil
}

func (f *fakeProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

// TestRunToolLoop proves the infer -> act -> feed-back cycle end to end with a real
// tool (bash) and a deterministic model: the model asks to run bash, the loop runs it,
// feeds the output back, and the model's final reply is returned.
func TestRunToolLoop(t *testing.T) {
	fp := &fakeProvider{}
	a := New(Config{
		Provider: fp, Tools: tools.Default(t.TempDir()),
		Cortex: cortex.NewMemoryStore(), Model: "m",
	})

	reply, trace, err := a.RunTrace(context.Background(), "do it")
	if err != nil {
		t.Fatalf("RunTrace: %v", err)
	}
	if reply != "all done" {
		t.Errorf("reply = %q, want %q", reply, "all done")
	}
	if len(trace) != 1 || trace[0].Name != "bash" {
		t.Fatalf("trace = %+v", trace)
	}
	if !strings.Contains(trace[0].Result, "MIMIRX") {
		t.Errorf("bash result = %q, want it to contain MIMIRX", trace[0].Result)
	}
	if trace[0].Err != "" {
		t.Errorf("unexpected tool error: %q", trace[0].Err)
	}
	// The second model call must include the tool result message.
	if len(fp.seen) < 2 {
		t.Fatalf("expected 2 model calls, got %d", len(fp.seen))
	}
	found := false
	for _, m := range fp.seen[1] {
		if m.Role == "tool" && strings.Contains(m.Content, "MIMIRX") {
			found = true
		}
	}
	if !found {
		t.Errorf("tool result was not fed back to the model: %+v", fp.seen[1])
	}
}

// TestRunUnknownTool proves an unknown tool call returns an error result to the
// model instead of crashing the loop.
func TestRunUnknownTool(t *testing.T) {
	fp := &fakeProvider{}
	fp.calls = 0
	// Override: make the first call ask for a tool that isn't registered.
	a := New(Config{
		Provider: &oneCallProvider{call: llm.ToolCall{ID: "x", Name: "nope", Arguments: `{}`}},
		Tools:    tools.Default(t.TempDir()), Cortex: cortex.NewMemoryStore(), Model: "m",
	})
	reply, trace, err := a.RunTrace(context.Background(), "go")
	if err != nil {
		t.Fatalf("RunTrace: %v", err)
	}
	if reply != "final" {
		t.Errorf("reply = %q", reply)
	}
	if len(trace) != 1 || trace[0].Err == "" {
		t.Errorf("expected an error step for unknown tool, got %+v", trace)
	}
	_ = fp
}

// oneCallProvider returns one tool call then final text.
type oneCallProvider struct {
	call  llm.ToolCall
	calls int
}

func (p *oneCallProvider) ID() string { return "one" }
func (p *oneCallProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	p.calls++
	if p.calls == 1 {
		return llm.GenerateResponse{ToolCalls: []llm.ToolCall{p.call}, FinishReason: "tool_calls"}, nil
	}
	return llm.GenerateResponse{Content: "final", FinishReason: "stop"}, nil
}
func (p *oneCallProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}
