package agent

import (
	"strings"
	"testing"
)

func TestDetectWebCreationTask(t *testing.T) {
	if !DetectWebCreationTask("Build a Next.js landing page for a local business with Google Maps SEO") {
		t.Fatal("expected DetectWebCreationTask to return true for web/SEO task")
	}
	if !DetectWebCreationTask("Create a Flutter UI dashboard and add llms.txt + JSON-LD Schema.org") {
		t.Fatal("expected DetectWebCreationTask to return true for Flutter/Schema.org task")
	}
	if DetectWebCreationTask("Refactor the mutex lock in internal/cortex/sqlite.go") {
		t.Fatal("expected DetectWebCreationTask to return false for pure backend mutex task")
	}
}

func TestBuildHarnessRulesBlockZeroBloat(t *testing.T) {
	// 1. Backend Implementation task: loads ONLY Step 4 Implementation micro-prompt (<350 chars), NO Web/SEO bloat.
	backendRules := BuildHarnessRulesBlock("Fix a race condition in the Go channel worker", false)
	if !strings.Contains(backendRules, "SDLC 4/7: IMPLEMENTATION") {
		t.Errorf("expected Step 4 IMPLEMENTATION micro-prompt, got: %s", backendRules)
	}
	if strings.Contains(backendRules, "Web/SEO") || strings.Contains(backendRules, "SDLC 1/7") {
		t.Errorf("did not expect Web/SEO or other SDLC phases in backend prompt, got: %s", backendRules)
	}
	if len(backendRules) > 400 {
		t.Errorf("expected compact phase-scoped prompt (<400 chars), got %d chars: %s", len(backendRules), backendRules)
	}

	// 2. Web Implementation task: loads Step 4 Implementation micro-prompt + compact 1-line Web/SEO Build overlay.
	webRules := BuildHarnessRulesBlock("Build a SaaS landing page with local SEO and Google Maps GBP schema", false)
	if !strings.Contains(webRules, "SDLC 4/7: IMPLEMENTATION") {
		t.Errorf("expected Step 4 IMPLEMENTATION in web build prompt, got: %s", webRules)
	}
	if !strings.Contains(webRules, "[Web/SEO Build]") || !strings.Contains(webRules, "llms.txt") || !strings.Contains(webRules, "[vendor claim]") {
		t.Errorf("expected compact [Web/SEO Build] overlay in web prompt, got: %s", webRules)
	}

	// 3. Planning task: loads ONLY Step 1 Planning micro-prompt.
	planRules := BuildHarnessRulesBlock("Define the project plan and scope charter", true)
	if !strings.Contains(planRules, "SDLC 1/7: PLANNING") {
		t.Errorf("expected Step 1 PLANNING micro-prompt, got: %s", planRules)
	}
}

func TestEvaluateStopGate(t *testing.T) {
	trace := []Step{
		{Name: "edit_block", Args: `{"path":"main.go"}`},
	}
	todos := []string{"- #1 [pending] Run unit tests"}
	gate := EvaluateStopGate(trace, todos)
	if !gate.HasPendingTodos || gate.PendingCount != 1 {
		t.Errorf("expected 1 pending todo, got %+v", gate)
	}
	if gate.RanVerification {
		t.Errorf("expected RanVerification=false when edit_block ran without run_shell test/vet")
	}
	if !strings.Contains(gate.Advisory, "pending todo") || !strings.Contains(gate.Advisory, "verification") {
		t.Errorf("unexpected advisory: %s", gate.Advisory)
	}

	traceVerified := []Step{
		{Name: "edit_block", Args: `{"path":"main.go"}`},
		{Name: "run_shell", Args: `{"command":"go test ./..."}`},
	}
	todosDone := []string{"- #1 [completed] Run unit tests"}
	gateOk := EvaluateStopGate(traceVerified, todosDone)
	if gateOk.HasPendingTodos || !gateOk.RanVerification || gateOk.Advisory != "" {
		t.Errorf("expected clean StopGateStatus, got %+v", gateOk)
	}
}
