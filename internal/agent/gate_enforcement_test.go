package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/policy"
	"github.com/David2024patton/Mimir/internal/sdlc"
	"github.com/David2024patton/Mimir/internal/tools"
)

func writeCall() llm.ToolCall {
	return llm.ToolCall{ID: "w1", Name: "write_file", Arguments: `{"path":"out.txt","content":"hi"}`}
}

// TestGateRequireApprovalNilApproverDenies proves the F2/N1 fix: a gate
// RequireApproval verdict with no approver configured fails closed instead
// of silently passing the tool call.
func TestGateRequireApprovalNilApproverDenies(t *testing.T) {
	dir := t.TempDir()
	a := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Gate: policy.DefaultDevGate(),
	})
	st := a.executeToolCall(context.Background(), writeCall())
	if st.Err == "" {
		t.Fatalf("expected denial with nil approver, got step %+v", st)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.txt")); !os.IsNotExist(err) {
		t.Errorf("write_file executed despite the missing approver")
	}
}

// TestGateRequireApprovalApproverDecides proves an approver returning false
// denies and an approver returning true allows the gated tool call.
func TestGateRequireApprovalApproverDecides(t *testing.T) {
	dir := t.TempDir()
	deny := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Gate:         policy.DefaultDevGate(),
		ApprovalFunc: func(string, string) bool { return false },
	})
	if st := deny.executeToolCall(context.Background(), writeCall()); st.Err == "" {
		t.Errorf("approver=false: expected denial, got %+v", st)
	}

	allow := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Gate:         policy.DefaultDevGate(),
		ApprovalFunc: func(string, string) bool { return true },
	})
	st := allow.executeToolCall(context.Background(), writeCall())
	if st.Err != "" {
		t.Fatalf("approver=true: unexpected error %q", st.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.txt")); err != nil {
		t.Errorf("approver=true: file was not written: %v", err)
	}
}

// TestBypassWaivesApprovalButNotDeny proves bypass mode is an explicit user
// opt-out of approval prompts only: RequireApproval is waived with no approver
// at all, while a Deny verdict still blocks execution.
func TestBypassWaivesApprovalButNotDeny(t *testing.T) {
	gate := policy.NewGate()
	gate.AddRule(policy.Rule{Pattern: "write_file", Decision: policy.RequireApproval})
	gate.AddRule(policy.Rule{Pattern: "bash", Decision: policy.Deny})

	dir := t.TempDir()
	a := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeBypass, Gate: gate,
	})
	st := a.executeToolCall(context.Background(), writeCall())
	if st.Err != "" {
		t.Errorf("bypass + RequireApproval: unexpected error %q", st.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.txt")); err != nil {
		t.Errorf("bypass + RequireApproval: file was not written: %v", err)
	}

	st = a.executeToolCall(context.Background(), llm.ToolCall{ID: "b1", Name: "bash", Arguments: `{"command":"echo hi"}`})
	if st.Err == "" || !strings.Contains(st.Err, "deny") {
		t.Errorf("bypass + Deny: expected a deny error, got %+v", st)
	}
}

// TestEditModeNilApproverDeniesWrite proves the ModeEdit approval branch also
// fails closed: in edit mode a write tool with no approver is denied even
// without a policy gate installed.
func TestEditModeNilApproverDeniesWrite(t *testing.T) {
	dir := t.TempDir()
	a := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeEdit,
	})
	st := a.executeToolCall(context.Background(), writeCall())
	if st.Err == "" {
		t.Fatalf("expected edit-mode denial with nil approver, got %+v", st)
	}
	if _, err := os.Stat(filepath.Join(dir, "out.txt")); !os.IsNotExist(err) {
		t.Errorf("write_file executed in edit mode despite the missing approver")
	}
}

// TestPhaseEnforcementAtExecution proves the F5 fix: under waterfall, a tool
// outside the active phase's allowed set is rejected inside executeToolCall,
// not merely hidden from the advertised schemas.
func TestPhaseEnforcementAtExecution(t *testing.T) {
	dir := t.TempDir()
	mgr := sdlc.NewManager(dir)
	if err := mgr.SetMethodology("waterfall"); err != nil {
		t.Fatalf("SetMethodology: %v", err)
	}
	if _, err := mgr.SetPhase("planning"); err != nil {
		t.Fatalf("SetPhase: %v", err)
	}
	a := New(Config{
		Provider: &oneCallProvider{}, Tools: tools.Default(dir),
		Cortex: cortex.NewMemoryStore(), Model: "m",
		Mode: ModeBypass, SDLC: mgr,
	})

	st := a.executeToolCall(context.Background(), llm.ToolCall{ID: "b1", Name: "bash", Arguments: `{"command":"echo hi"}`})
	if st.Err == "" || !strings.Contains(st.Err, "SDLC phase policy") {
		t.Errorf("bash in waterfall planning phase: expected a phase policy rejection, got %+v", st)
	}

	// read_file IS in the planning phase set, so it must not be phase-blocked
	// (it errors only because the file does not exist).
	st = a.executeToolCall(context.Background(), llm.ToolCall{ID: "r1", Name: "read_file", Arguments: `{"path":"missing.txt"}`})
	if strings.Contains(st.Err, "SDLC phase policy") {
		t.Errorf("read_file in planning phase was phase-blocked: %q", st.Err)
	}
}
