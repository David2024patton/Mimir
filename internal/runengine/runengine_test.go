package runengine

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/David2024patton/Mimir/internal/sdlc"
	"github.com/David2024patton/Mimir/internal/tools"
)

type fakeExecutor struct {
	steps        []StepResult
	stepIdx      int
	targetedCall []string
	fullCalled   int
	fullPass     bool
}

func (f *fakeExecutor) RunStep(_ context.Context, _ string) (StepResult, error) {
	if f.stepIdx >= len(f.steps) {
		return StepResult{Reply: "done", Done: true}, nil
	}
	res := f.steps[f.stepIdx]
	f.stepIdx++
	return res, nil
}

func (f *fakeExecutor) RunTargetedTests(_ context.Context, targets []string) (TestRunResult, error) {
	f.targetedCall = append([]string(nil), targets...)
	return TestRunResult{
		Tier:    "targeted_inner_loop",
		Passed:  true,
		Targets: targets,
		Output:  "ok targeted",
	}, nil
}

func (f *fakeExecutor) RunFullSuite(_ context.Context) (TestRunResult, error) {
	f.fullCalled++
	if !f.fullPass {
		return TestRunResult{
			Tier:    "full_suite_gate",
			Passed:  false,
			Targets: []string{"./..."},
			Output:  "FAIL ./internal/broken",
		}, errors.New("exit status 1")
	}
	return TestRunResult{
		Tier:    "full_suite_gate",
		Passed:  true,
		Targets: []string{"./..."},
		Output:  "ok ./...",
	}, nil
}

func TestJailBlocksPathTraversalAndProtectedPaths(t *testing.T) {
	dir := t.TempDir()
	jail, err := NewJail(dir)
	if err != nil {
		t.Fatalf("NewJail: %v", err)
	}

	// Valid write and read inside workspace jail
	if err := jail.WriteFile("pkg/hello.go", []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("expected valid write inside jail, got %v", err)
	}
	got, err := jail.ReadFile("pkg/hello.go")
	if err != nil || string(got) != "package pkg\n" {
		t.Fatalf("ReadFile mismatch: %q, err=%v", string(got), err)
	}
	entries, err := jail.ListDir(".")
	if err != nil || len(entries) == 0 {
		t.Fatalf("ListDir failed: %v (%v)", entries, err)
	}

	// Reject parent traversal
	if _, err := jail.Resolve("../outside.txt"); !errors.Is(err, ErrJailEscape) {
		t.Fatalf("expected ErrJailEscape for ../outside.txt, got %v", err)
	}
	if err := jail.WriteFile("../../escape.sh", []byte("echo nope"), 0o644); !errors.Is(err, ErrJailEscape) {
		t.Fatalf("expected ErrJailEscape on write, got %v", err)
	}

	// Reject protected .git and .mimir/run.lock writes
	if err := jail.WriteFile(".git/config", []byte("bad"), 0o644); !errors.Is(err, ErrProtectedPath) {
		t.Fatalf("expected ErrProtectedPath for .git/config, got %v", err)
	}
	if err := jail.WriteFile(".mimir/run.lock", []byte("bad"), 0o644); !errors.Is(err, ErrProtectedPath) {
		t.Fatalf("expected ErrProtectedPath for .mimir/run.lock, got %v", err)
	}
}

func TestRunLockConcurrencyAndStaleRecovery(t *testing.T) {
	dir := t.TempDir()

	lock1, err := AcquireLock(dir, "task-1", 10*time.Minute)
	if err != nil {
		t.Fatalf("AcquireLock 1 failed: %v", err)
	}

	// Second concurrent lock must fail with ErrRunLocked
	if _, err := AcquireLock(dir, "task-2", 10*time.Minute); !errors.Is(err, ErrRunLocked) {
		t.Fatalf("expected ErrRunLocked on active lock, got %v", err)
	}

	// Release and re-acquire
	if err := lock1.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	lock2, err := AcquireLock(dir, "task-2", 10*time.Minute)
	if err != nil {
		t.Fatalf("AcquireLock after release failed: %v", err)
	}
	_ = lock2.Release()

	// Write a stale lock older than StaleAfter and verify automatic recovery
	staleMeta := LockMetadata{
		PID:       999999,
		Task:      "crashed-task",
		StartedAt: time.Now().UTC().Add(-2 * time.Hour),
	}
	raw, _ := json.Marshal(staleMeta)
	lockPath := filepath.Join(dir, ".mimir", "run.lock")
	if err := os.WriteFile(lockPath, raw, 0o600); err != nil {
		t.Fatalf("write stale lock: %v", err)
	}

	recovered, err := AcquireLock(dir, "fresh-task", 30*time.Minute)
	if err != nil {
		t.Fatalf("expected stale lock recovery, got %v", err)
	}
	defer func() { _ = recovered.Release() }()
}

func TestTwoTierTestingLaw_InnerTargetedVsOuterFullSuite(t *testing.T) {
	dir := t.TempDir()
	fx := &fakeExecutor{fullPass: true}
	eng, err := New(dir, fx, DefaultLimits())
	if err != nil {
		t.Fatalf("New engine: %v", err)
	}

	candidates := []string{
		"internal/sdlc/sdlc_test.go",
		"internal/critic/epistemic_test.go",
		"internal/server/server_test.go",
	}

	// Tier 1: Inner edit loop selects only affected tests via Huginn-Decider
	innerRes, err := eng.RunInnerLoopTests(context.Background(), []string{"internal/critic/epistemic.go"}, candidates, 1)
	if err != nil {
		t.Fatalf("RunInnerLoopTests: %v", err)
	}
	if innerRes.Tier != "targeted_inner_loop" {
		t.Fatalf("expected targeted_inner_loop tier, got %s", innerRes.Tier)
	}
	if len(fx.targetedCall) != 1 || fx.targetedCall[0] != "internal/critic/epistemic_test.go" {
		t.Fatalf("expected Huginn-Decider to pick internal/critic/epistemic_test.go, got %v", fx.targetedCall)
	}
	if fx.fullCalled != 0 {
		t.Fatalf("inner loop must not run full suite, fullCalled=%d", fx.fullCalled)
	}

	// Tier 2: ReviewGate always runs full test suite
	gate := eng.ReviewGate(context.Background(), GateOptions{})
	if !gate.Passed {
		t.Fatalf("expected ReviewGate to pass, got %+v", gate)
	}
	if fx.fullCalled != 1 || gate.FullSuiteResult.Tier != "full_suite_gate" {
		t.Fatalf("expected ReviewGate to run full suite once, fullCalled=%d tier=%s", fx.fullCalled, gate.FullSuiteResult.Tier)
	}
}

func TestReviewAndShipGate_ChecklistSDLCAndEpistemicEnforcement(t *testing.T) {
	dir := t.TempDir()
	fx := &fakeExecutor{fullPass: true}
	eng, err := New(dir, fx, DefaultLimits())
	if err != nil {
		t.Fatalf("New engine: %v", err)
	}

	// 1. Pending todo blocks gate
	td := eng.Todos.Add("Build runengine")
	rep := eng.ReviewGate(context.Background(), GateOptions{})
	if rep.Passed || rep.ChecklistPassed {
		t.Fatalf("expected pending todo to block ReviewGate, got %+v", rep)
	}

	// Mark todo completed
	eng.Todos.SetStatus(td.ID, tools.TodoCompleted)

	// 2. Missing required SDLC deliverable blocks gate
	rep = eng.ReviewGate(context.Background(), GateOptions{
		RequiredPhases: []sdlc.PhaseID{sdlc.PhasePlanning},
	})
	if rep.Passed || rep.SDLCPassed || len(rep.MissingArtifacts) != 1 {
		t.Fatalf("expected missing 01-PLAN.md to block ReviewGate, got %+v", rep)
	}

	// Save the Planning deliverable
	if _, err := eng.SDLC.SaveDeliverable(sdlc.PhasePlanning, "Charter ready"); err != nil {
		t.Fatalf("SaveDeliverable: %v", err)
	}

	// 3. Unbacked epistemic claim blocks gate
	rep = eng.ReviewGate(context.Background(), GateOptions{
		RequiredPhases:   []sdlc.PhaseID{sdlc.PhasePlanning},
		RetrievedContext: "go test ./... passed in 12 packages.",
		DraftReport:      "Our engine is 10x faster and achieves 99.4% accuracy across all benchmarks.",
	})
	if rep.Passed || rep.EpistemicPassed {
		t.Fatalf("expected unbacked claims to block Epistemic Gate, got %+v", rep)
	}

	// 4. Grounded report + completed todos + deliverable + green full suite passes ShipGate
	shipRep := eng.ShipGate(context.Background(), GateOptions{
		RequiredPhases:   []sdlc.PhaseID{sdlc.PhasePlanning},
		RetrievedContext: "go test ./... passed in 12 packages.",
		DraftReport:      "go test ./... passed in 12 packages.",
	})
	if !shipRep.Passed {
		t.Fatalf("expected ShipGate to pass, got:\n%s", shipRep.FormatMarkdown())
	}
	if shipRep.SDLCPhase != string(sdlc.PhaseDeployment) {
		t.Fatalf("expected ShipGate to advance SDLC to deployment, got %s", shipRep.SDLCPhase)
	}
}

func TestRunUnattended_EnforcesLimitsAndLock(t *testing.T) {
	dir := t.TempDir()
	limits := DefaultLimits()
	limits.MaxSteps = 3
	limits.MaxTokens = 500

	fx := &fakeExecutor{
		fullPass: true,
		steps: []StepResult{
			{Reply: "edited critic", TokensUsed: 150, EditedFiles: []string{"internal/critic/epistemic.go"}, Done: false},
			{Reply: "finished", TokensUsed: 150, Done: true},
		},
	}
	eng, err := New(dir, fx, limits)
	if err != nil {
		t.Fatalf("New engine: %v", err)
	}

	summary, err := eng.RunUnattended(context.Background(), "update critic", []string{"internal/critic/epistemic_test.go"})
	if err != nil {
		t.Fatalf("RunUnattended failed: %v", err)
	}
	if summary.Steps != 2 || len(summary.InnerTests) != 1 || !summary.Gate.Passed {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	// Verify token budget enforcement
	fxOver := &fakeExecutor{
		fullPass: true,
		steps: []StepResult{
			{Reply: "huge step", TokensUsed: 900, Done: false},
		},
	}
	engOver, _ := New(dir, fxOver, limits)
	_, err = engOver.RunUnattended(context.Background(), "blow token budget", nil)
	if !errors.Is(err, ErrTokenLimitExceeded) {
		t.Fatalf("expected ErrTokenLimitExceeded, got %v", err)
	}
	if !strings.Contains(err.Error(), "900") {
		t.Fatalf("expected token usage in error, got %v", err)
	}
}
