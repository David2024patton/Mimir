package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// scriptProvider replays a fixed script of Generate responses in order.
type scriptProvider struct {
	responses []llm.GenerateResponse
	calls     int
}

func (p *scriptProvider) ID() string { return "script" }
func (p *scriptProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	p.calls++
	if p.calls <= len(p.responses) {
		return p.responses[p.calls-1], nil
	}
	return llm.GenerateResponse{Content: "script exhausted", FinishReason: "stop"}, nil
}
func (p *scriptProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

func toolResp(name, args string) llm.GenerateResponse {
	return llm.GenerateResponse{
		ToolCalls:    []llm.ToolCall{{ID: "s1", Name: name, Arguments: args}},
		FinishReason: "tool_calls",
	}
}

func textResp(text string) llm.GenerateResponse {
	return llm.GenerateResponse{Content: text, FinishReason: "stop"}
}

// TestAutoVerifyForcesVerificationPass proves the F25 fix: with AutoVerify on,
// a turn that mutates a file and stops without verification is not finished;
// the loop continues, and once the model runs a verification command the turn
// completes without an advisory.
func TestAutoVerifyForcesVerificationPass(t *testing.T) {
	dir := t.TempDir()
	prov := &scriptProvider{responses: []llm.GenerateResponse{
		toolResp("write_file", `{"path":"out.txt","content":"hi"}`),
		textResp("done, trust me"),
		toolResp("bash", `{"command":"echo check"}`),
		textResp("verified and done"),
	}}
	a := New(Config{
		Provider: prov, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeBypass, AutoVerify: true,
	})
	reply, _, err := a.RunTrace(context.Background(), "build it", nil)
	if err != nil {
		t.Fatalf("RunTrace: %v", err)
	}
	if prov.calls != 4 {
		t.Errorf("provider calls = %d, want 4 (the unverified stop must be forced to continue)", prov.calls)
	}
	if reply != "verified and done" {
		t.Errorf("reply = %q, want the post-verification reply with no advisory", reply)
	}
}

// TestAutoVerifyContinuationCap proves the forced continuations are capped at
// 2 per run: a model that never verifies still finishes, with the advisory
// attached, instead of looping forever.
func TestAutoVerifyContinuationCap(t *testing.T) {
	dir := t.TempDir()
	prov := &scriptProvider{responses: []llm.GenerateResponse{
		toolResp("write_file", `{"path":"out.txt","content":"hi"}`),
		textResp("done"),
		textResp("still done"),
		textResp("really done"),
	}}
	a := New(Config{
		Provider: prov, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeBypass, AutoVerify: true,
	})
	reply, _, err := a.RunTrace(context.Background(), "build it", nil)
	if err != nil {
		t.Fatalf("RunTrace: %v", err)
	}
	// 1 tool turn + 1 stop + 2 forced continuations = 4 Generate calls.
	if prov.calls != 4 {
		t.Errorf("provider calls = %d, want 4 (cap of 2 forced continuations)", prov.calls)
	}
	if !strings.Contains(reply, "Harness OnStop Gate Advisory") {
		t.Errorf("reply = %q, want the advisory attached after the cap", reply)
	}
}

// TestAutoVerifyOffKeepsAdvisoryOnly proves the pre-F25 behavior is preserved
// when AutoVerify is disabled: the gate annotates and the turn ends.
func TestAutoVerifyOffKeepsAdvisoryOnly(t *testing.T) {
	dir := t.TempDir()
	prov := &scriptProvider{responses: []llm.GenerateResponse{
		toolResp("write_file", `{"path":"out.txt","content":"hi"}`),
		textResp("done"),
	}}
	a := New(Config{
		Provider: prov, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeBypass, AutoVerify: false,
	})
	reply, _, err := a.RunTrace(context.Background(), "build it", nil)
	if err != nil {
		t.Fatalf("RunTrace: %v", err)
	}
	if prov.calls != 2 {
		t.Errorf("provider calls = %d, want 2 (no forced continuation when AutoVerify is off)", prov.calls)
	}
	if !strings.Contains(reply, "Harness OnStop Gate Advisory") {
		t.Errorf("reply = %q, want the advisory attached", reply)
	}
}
