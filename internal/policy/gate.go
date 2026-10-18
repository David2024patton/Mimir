package policy

import "strings"

// Decision is the result of evaluating a tool call against policy.
type Decision string

const (
	Allow           Decision = "allow"
	Deny            Decision = "deny"
	RequireApproval Decision = "require_approval"
)

// Request is a tool call to evaluate.
type Request struct {
	Tool string
	Args map[string]any
}

// Rule is one allow/deny/require_approval rule (pattern on tool/command).
type Rule struct {
	Pattern  string // tool name pattern or wildcard "*"
	Command  string // optional command pattern for shell/bash tools
	Decision Decision
}

// Gate is the fail-closed policy gate (E4.1, ADR-009).
type Gate struct {
	rules      []Rule
	defaultDec Decision
}

// NewGate returns a fail-closed gate (default-deny).
func NewGate() *Gate {
	return &Gate{defaultDec: Deny}
}

// DefaultDevGate returns a standard developer policy gate:
// read-only tools and harmless ops are Allowed; mutations & shell require approval or specific allow.
func DefaultDevGate() *Gate {
	g := NewGate()
	// Safe read-only inspections
	g.AddRule(Rule{Pattern: "read_file", Decision: Allow})
	g.AddRule(Rule{Pattern: "list_directory", Decision: Allow})
	g.AddRule(Rule{Pattern: "todo", Decision: Allow})
	g.AddRule(Rule{Pattern: "question", Decision: Allow})
	g.AddRule(Rule{Pattern: "load_skill", Decision: Allow})
	g.AddRule(Rule{Pattern: "epistemic_audit", Decision: Allow})
	g.AddRule(Rule{Pattern: "web_seo_audit", Decision: Allow})

	// Modifications require approval unless bypass mode or explicit rule
	g.AddRule(Rule{Pattern: "write_file", Decision: RequireApproval})
	g.AddRule(Rule{Pattern: "edit_block", Decision: RequireApproval})
	g.AddRule(Rule{Pattern: "bash", Decision: RequireApproval})
	g.AddRule(Rule{Pattern: "shell", Decision: RequireApproval})
	return g
}

func (g *Gate) AddRule(r Rule) {
	g.rules = append(g.rules, r)
}

// Evaluate returns the decision for a request (default-deny; last matching rule wins).
func (g *Gate) Evaluate(req Request) Decision {
	decision := g.defaultDec
	for _, r := range g.rules {
		if match(r.Pattern, req.Tool) {
			if r.Command != "" && (req.Tool == "bash" || req.Tool == "shell") {
				cmd, _ := req.Args["command"].(string)
				if match(r.Command, cmd) {
					decision = r.Decision
				}
			} else {
				decision = r.Decision
			}
		}
	}
	return decision
}

func match(pattern, s string) bool {
	if pattern == "*" || pattern == s {
		return true
	}
	// Case-insensitive substring match when pattern starts with contains:
	if len(pattern) > 9 && pattern[:9] == "contains:" {
		return strings.Contains(strings.ToLower(s), strings.ToLower(pattern[9:]))
	}
	return false
}
