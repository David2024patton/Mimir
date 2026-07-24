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
