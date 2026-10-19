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

// TestSteerReportsRejection proves Steer never silently drops: empty session
// IDs and empty messages are rejected, a full per-session queue rejects, and
// the queue accepts again after a drain. (Pre-fix, Steer returned nothing
// and a non-blocking channel send dropped overflow without a trace.)
func TestSteerReportsRejection(t *testing.T) {
	a := New(Config{Provider: &fakeProvider{}, Tools: tools.NewRegistry(), Cortex: cortex.NewMemoryStore(), Model: "m"})

	if a.Steer("", "no session") {
		t.Error("Steer with empty session ID reported success")
	}
	if a.Steer("sess", "") {
		t.Error("Steer with empty message reported success")
	}
	if a.Steer("sess", "   ") {
		t.Error("Steer with whitespace-only message reported success")
	}
	for i := 0; i < maxSteersPerSession; i++ {
		if !a.Steer("sess", "queued") {
			t.Fatalf("Steer %d of %d reported failure below the cap", i+1, maxSteersPerSession)
		}
	}
	if a.Steer("sess", "one too many") {
		t.Error("Steer beyond the per-session cap reported success")
	}
	// A different session has its own queue and is unaffected by the full one.
	if !a.Steer("other", "still fine") {
		t.Error("Steer for a different session was rejected by another session's full queue")
	}
	if got := a.drainSteer("sess"); len(got) != maxSteersPerSession {
		t.Fatalf("drainSteer(sess) returned %d messages, want %d", len(got), maxSteersPerSession)
	}
	if !a.Steer("sess", "after drain") {
		t.Error("Steer after drain reported failure")
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
