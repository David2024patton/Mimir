package decision

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeciderChoiceAndPruneTools(t *testing.T) {
	d := New("", "", "")
	allTools := []string{
		"read_file", "edit_block", "todo", "bash", "web_search", "grep_search", "load_skill", "question",
	}

	pruned := d.PruneToolNames(context.Background(), "run bash tests and grep_search for errors", allTools, 5)
	if len(pruned) != 5 {
		t.Fatalf("got %d tools, want 5: %v", len(pruned), pruned)
	}
	hasBash := false
	for _, name := range pruned {
		if name == "bash" {
			hasBash = true
		}
	}
	if !hasBash {
		t.Errorf("expected pruned tools to include bash, got %v", pruned)
	}
}

func TestDeciderRoutePrompt(t *testing.T) {
	d := New("", "", "")

	r1 := d.RoutePrompt(context.Background(), "Design and refactor the polyglot system architecture")
	if r1.Persona != "Odin" || r1.Tier != "T1" || !r1.NeedsThinking {
		t.Errorf("unexpected route for architecture prompt: %+v", r1)
	}

	r2 := d.RoutePrompt(context.Background(), "Fix race condition panic in worker loop")
	if r2.Persona != "Loki" || r2.Tier != "T1" {
		t.Errorf("unexpected route for bug prompt: %+v", r2)
	}

	r3 := d.RoutePrompt(context.Background(), "Rename variable foo to bar")
	if r3.Persona != "Thor" || r3.Tier != "T3" || r3.NeedsThinking {
		t.Errorf("unexpected route for simple edit: %+v", r3)
	}
}

func TestDeciderAssessCommandSafety(t *testing.T) {
	d := New("", "", "")

	safe := d.AssessCommandSafety(context.Background(), "go test ./...")
	if !safe.AutoApprove || safe.SafeScore < 0.9 {
		t.Errorf("expected go test to auto-approve, got %+v", safe)
	}

	risky := d.AssessCommandSafety(context.Background(), "git reset --hard HEAD~5")
	if risky.AutoApprove || risky.SafeScore > 0.2 {
		t.Errorf("expected git reset --hard to be blocked from auto-approve, got %+v", risky)
	}
}

func TestEnsureLocalDecisionModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"models":[{"name":"liquid/d1-3b:latest"}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	model, err := EnsureLocalDecisionModel(context.Background(), srv.URL, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "liquid/d1-3b:latest" {
		t.Errorf("got %q, want liquid/d1-3b:latest", model)
	}
}
