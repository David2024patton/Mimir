package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TestSteerQueuesArePerSession proves the queue is keyed by session: a steer
// queued for one session is invisible to every other session, a sessionless
// drain gets nothing, and a drain consumes the queue exactly once.
func TestSteerQueuesArePerSession(t *testing.T) {
	a := New(Config{Provider: &fakeProvider{}, Tools: tools.NewRegistry(), Cortex: cortex.NewMemoryStore(), Model: "m"})

	if !a.Steer("sess-a", "for a only") {
		t.Fatal("Steer for sess-a was rejected")
	}
	if got := a.drainSteer("sess-b"); len(got) != 0 {
		t.Errorf("drainSteer(sess-b) = %v, want empty (per-session isolation)", got)
	}
	if got := a.drainSteer(""); len(got) != 0 {
		t.Errorf("drainSteer(\"\") = %v, want empty (sessionless runs drain nothing)", got)
	}
	got := a.drainSteer("sess-a")
	if len(got) != 1 || got[0] != "for a only" {
		t.Fatalf("drainSteer(sess-a) = %v, want [for a only]", got)
	}
	if got := a.drainSteer("sess-a"); len(got) != 0 {
		t.Errorf("second drainSteer(sess-a) = %v, want empty (queue consumed)", got)
	}
}

// steerLoopProvider is run-aware like the server probe: it returns one tool
// call when the request does not yet end with a tool result, then final text,
// and records every message across all runs.
type steerLoopProvider struct {
	seen [][]llm.Message
}

func (p *steerLoopProvider) ID() string { return "steerloop" }
func (p *steerLoopProvider) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	p.seen = append(p.seen, req.Messages)
	if n := len(req.Messages); n == 0 || req.Messages[n-1].Role != "tool" {
		return llm.GenerateResponse{
			ToolCalls:    []llm.ToolCall{{ID: "s1", Name: "todo", Arguments: "{}"}},
			FinishReason: "tool_calls",
		}, nil
	}
	return llm.GenerateResponse{Content: "done", FinishReason: "stop"}, nil
}
func (p *steerLoopProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}
func (p *steerLoopProvider) sawSteering(msg string) bool {
	for _, call := range p.seen {
		for _, m := range call {
			if strings.Contains(m.Content, "[User steering instruction]: "+msg) {
				return true
			}
		}
	}
	return false
}

// TestRunDeliversSteerOnlyToBoundSession proves the end-to-end binding at
// the agent level: a steer queued for session X is delivered only into a run
// whose context carries session X (WithSessionID). A run bound to another
// session and a sessionless run receive nothing, and delivery happens once.
func TestRunDeliversSteerOnlyToBoundSession(t *testing.T) {
	prov := &steerLoopProvider{}
	a := New(Config{
		Provider: prov, Tools: tools.NewRegistry(), Cortex: cortex.NewMemoryStore(), Model: "m",
		ApprovalFunc: func(string, string) bool { return true },
	})
	if !a.Steer("sess-x", "bound steer") {
		t.Fatal("Steer for sess-x was rejected")
	}
	run := func(ctx context.Context) {
		t.Helper()
		if _, err := a.Run(ctx, "go", nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
	}

	run(WithSessionID(context.Background(), "sess-y"))
	if prov.sawSteering("bound steer") {
		t.Fatal("steer for sess-x leaked into a run bound to sess-y")
	}
	run(context.Background())
	if prov.sawSteering("bound steer") {
		t.Fatal("steer for sess-x leaked into a sessionless run")
	}
	run(WithSessionID(context.Background(), "sess-x"))
	if !prov.sawSteering("bound steer") {
		t.Fatal("steer for sess-x was not delivered into the sess-x run")
	}
}
