package agent

import (
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/decision"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TestDefaultDeciderEndpoint proves the F26b fix: the default decider is no
// longer constructed with a permanently empty base URL. It inherits the main
// provider's base URL, and MIMIR_DECISION_BASE_URL overrides it.
func TestDefaultDeciderEndpoint(t *testing.T) {
	prov := &llm.OpenAIProvider{IDStr: "p", BaseURL: "http://127.0.0.1:11434/v1", Model: "m"}
	a := New(Config{
		Provider: prov, Tools: tools.NewRegistry(),
		Cortex: cortex.NewMemoryStore(), Model: "m",
	})
	if got := a.Decider().BaseURL; got != "http://127.0.0.1:11434/v1" {
		t.Errorf("decider BaseURL = %q, want the provider base URL", got)
	}
	if got := a.Decider().Model; got != decision.PrimaryDecisionModel {
		t.Errorf("decider Model = %q, want %q", got, decision.PrimaryDecisionModel)
	}

	t.Setenv("MIMIR_DECISION_BASE_URL", "http://127.0.0.1:9999/v1")
	t.Setenv("MIMIR_DECISION_MODEL", "d1-test")
	a = New(Config{
		Provider: prov, Tools: tools.NewRegistry(),
		Cortex: cortex.NewMemoryStore(), Model: "m",
	})
	if got := a.Decider().BaseURL; got != "http://127.0.0.1:9999/v1" {
		t.Errorf("decider BaseURL = %q, want the env override", got)
	}
	if got := a.Decider().Model; got != "d1-test" {
		t.Errorf("decider Model = %q, want the env override", got)
	}
}
