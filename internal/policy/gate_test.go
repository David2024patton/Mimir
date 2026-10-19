package policy

import "testing"

func TestGateFailClosedByDefault(t *testing.T) {
	g := NewGate()
	if d := g.Evaluate(Request{Tool: "bash"}); d != Deny {
		t.Errorf("default decision = %q, want deny (fail-closed)", d)
	}
}

func TestGateAllowRule(t *testing.T) {
	g := NewGate()
	g.AddRule(Rule{Pattern: "read_file", Decision: Allow})
	if d := g.Evaluate(Request{Tool: "read_file"}); d != Allow {
		t.Errorf("decision = %q, want allow", d)
	}
	if d := g.Evaluate(Request{Tool: "bash"}); d != Deny {
		t.Errorf("unmatched decision = %q, want deny", d)
	}
}

func TestGateWildcard(t *testing.T) {
	g := NewGate()
	g.AddRule(Rule{Pattern: "*", Decision: RequireApproval})
	if d := g.Evaluate(Request{Tool: "anything"}); d != RequireApproval {
		t.Errorf("wildcard decision = %q, want require_approval", d)
	}
}

func TestGateCommandContentMatching(t *testing.T) {
	g := NewGate()
	// Disallow rm -rf or destructive commands
	g.AddRule(Rule{Pattern: "bash", Command: "contains:rm -rf", Decision: Deny})
	// Allow safe test commands
	g.AddRule(Rule{Pattern: "bash", Command: "contains:go test", Decision: Allow})

	if d := g.Evaluate(Request{Tool: "bash", Args: map[string]any{"command": "go test ./..."}}); d != Allow {
		t.Errorf("safe command decision = %q, want allow", d)
	}
	if d := g.Evaluate(Request{Tool: "bash", Args: map[string]any{"command": "rm -rf /"}}); d != Deny {
		t.Errorf("destructive command decision = %q, want deny", d)
	}
}

func TestDefaultDevGate(t *testing.T) {
	g := DefaultDevGate()
	if d := g.Evaluate(Request{Tool: "read_file"}); d != Allow {
		t.Errorf("read_file decision = %q, want allow", d)
	}
	if d := g.Evaluate(Request{Tool: "write_file"}); d != RequireApproval {
		t.Errorf("write_file decision = %q, want require_approval", d)
	}
	if d := g.Evaluate(Request{Tool: "bash"}); d != RequireApproval {
		t.Errorf("bash decision = %q, want require_approval", d)
	}
	if d := g.Evaluate(Request{Tool: "unknown_dangerous"}); d != Deny {
		t.Errorf("unregistered tool decision = %q, want deny (fail-closed)", d)
	}
}

// TestDefaultDevGateCoversRegisteredTools proves the gate regression is fixed:
// the sdlc and ingest_repo tools (registered in tools.Default) must not fall
// through to default-deny in the shipped gate.
func TestDefaultDevGateCoversRegisteredTools(t *testing.T) {
	g := DefaultDevGate()
	for _, tool := range []string{"sdlc", "ingest_repo"} {
		if d := g.Evaluate(Request{Tool: tool}); d == Deny {
			t.Errorf("%s decision = deny, want non-deny (tool must stay reachable)", tool)
		}
	}
}
