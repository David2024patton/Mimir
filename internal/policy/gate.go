package policy

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
	Pattern  string
	Decision Decision
}

// Gate is the fail-closed policy gate (E4.1, ADR-009).
type Gate struct {
	rules       []Rule
	defaultDec  Decision
}

// NewGate returns a fail-closed gate (default-deny).
func NewGate() *Gate {
	return &Gate{defaultDec: Deny}
}

func (g *Gate) AddRule(r Rule) {
	g.rules = append(g.rules, r)
}

// Evaluate returns the decision for a request (default-deny; last matching rule wins).
func (g *Gate) Evaluate(req Request) Decision {
	decision := g.defaultDec
	for _, r := range g.rules {
		if match(r.Pattern, req.Tool) {
			decision = r.Decision
		}
	}
	return decision
}

func match(pattern, s string) bool {
	// TODO(E4.1/E4.2): glob matching + shell-AST parsing.
	return pattern == "*" || pattern == s
}
