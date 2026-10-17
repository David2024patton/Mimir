package sdlc

import (
	"strings"
	"testing"
)

func TestAllPhasesCompactNoBloat(t *testing.T) {
	phases := AllPhases()
	if len(phases) != 7 {
		t.Fatalf("expected 7 SDLC phases, got %d", len(phases))
	}
	for i, p := range phases {
		if p.Step != i+1 {
			t.Errorf("phase %s step = %d, want %d", p.ID, p.Step, i+1)
		}
		// Ensure micro-prompts are tiny (<300 chars) so neither big nor tiny LLMs suffer prompt bloat.
		if len(p.MicroPrompt) > 300 {
			t.Errorf("phase %s MicroPrompt too bloated (%d chars): %s", p.ID, len(p.MicroPrompt), p.MicroPrompt)
		}
		if len(p.AllowedTools) == 0 {
			t.Errorf("phase %s has empty AllowedTools", p.ID)
		}
	}
}

func TestDetectPhaseFromPrompt(t *testing.T) {
	cases := []struct {
		prompt string
		want   PhaseID
	}{
		{"Define the project plan, scope, and feasibility charter", PhasePlanning},
		{"Gather the SRS requirements and analyze repo constraints", PhaseAnalysis},
		{"Blueprint the system design architecture and database schema", PhaseDesign},
		{"Add a new handler function in internal/server/server.go", PhaseImplementation},
		{"Run unit tests and linter to verify quality", PhaseTesting},
		{"Deploy the production container via CI/CD pipeline", PhaseDeployment},
		{"Monitor telemetry logs and patch vulnerability", PhaseMaintenance},
	}
	for _, tc := range cases {
		if got := DetectPhaseFromPrompt(tc.prompt); got != tc.want {
			t.Errorf("DetectPhaseFromPrompt(%q) = %q, want %q", tc.prompt, got, tc.want)
		}
	}
}

func TestWaterfallGateAndSaveDeliverable(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.SetMethodology("waterfall"); err != nil {
		t.Fatalf("SetMethodology: %v", err)
	}
	if _, err := mgr.SetPhase("1"); err != nil {
		t.Fatalf("SetPhase(1): %v", err)
	}
	// Advancing in Waterfall without signing off Step 1 must fail.
	if _, err := mgr.Advance(); err == nil {
		t.Fatal("expected Waterfall gate to block advancing before Step 1 sign-off")
	}
	// Saving the deliverable signs off Step 1 and writes .mimir/sdlc/01-PLAN.md.
	relPath, err := mgr.SaveDeliverable(PhasePlanning, "Scope: 4-stack polyglot harness.")
	if err != nil {
		t.Fatalf("SaveDeliverable: %v", err)
	}
	if relPath != ".mimir/sdlc/01-PLAN.md" {
		t.Errorf("relPath = %q, want .mimir/sdlc/01-PLAN.md", relPath)
	}
	next, err := mgr.Advance()
	if err != nil {
		t.Fatalf("Advance after sign-off failed: %v", err)
	}
	if next.ID != PhaseAnalysis {
		t.Errorf("next phase = %q, want %q", next.ID, PhaseAnalysis)
	}
}

func TestBuildPhasePromptLoadsOnlyActivePhase(t *testing.T) {
	mgr := NewManager(t.TempDir())
	// Backend testing prompt -> loads ONLY Step 5 Testing micro-prompt, no WebOverlay, no Planning/Deployment bloat.
	p := mgr.BuildPhasePrompt("Run unit tests to verify the package", false)
	if !strings.Contains(p, "SDLC 5/7: TESTING") {
		t.Errorf("expected Step 5 TESTING micro-prompt, got: %s", p)
	}
	if strings.Contains(p, "SDLC 1/7") || strings.Contains(p, "SDLC 3/7") || strings.Contains(p, "Web/SEO") {
		t.Errorf("expected zero bloat from other phases or web overlay, got: %s", p)
	}

	// Web design prompt -> loads ONLY Step 3 Design micro-prompt + compact 1-line WebOverlay.
	w := mgr.BuildPhasePrompt("Design the system architecture and wireframe for our SaaS website", true)
	if !strings.Contains(w, "SDLC 3/7: DESIGN") || !strings.Contains(w, "[Web/SEO Design]") {
		t.Errorf("expected Step 3 DESIGN + compact Web/SEO Design overlay, got: %s", w)
	}
}
