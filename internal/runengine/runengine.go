// Package runengine implements Mimir's unattended execution engine:
//   - Shell-less workspace file jail (prevents path traversal, symlink escapes, and .git tampering)
//   - Single-run lockfile (.mimir/run.lock) with stale lock detection
//   - Configurable execution limits (MaxSteps, MaxTokens, Timeout)
//   - Executor seam interface for deterministic testing and shell-free argv execution
//   - Explicit Two-Tier Testing Law:
//     * Tier 1 (Inner Edit Loop, SDLC Steps 4-5): Huginn-Decider selects affected tests for fast feedback.
//     * Tier 2 (Outer Ship Gate, mimir review / mimir ship): Always runs the full test suite + Checklist Gate + Epistemic Critic.
package runengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/David2024patton/Mimir/internal/critic"
	"github.com/David2024patton/Mimir/internal/decision"
	"github.com/David2024patton/Mimir/internal/sdlc"
	"github.com/David2024patton/Mimir/internal/tools"
)

var (
	// ErrJailEscape is returned when a path resolves outside the workspace root.
	ErrJailEscape = errors.New("runengine: path escapes workspace jail")
	// ErrProtectedPath is returned when an unattended write targets .git or .mimir/run.lock.
	ErrProtectedPath = errors.New("runengine: write to protected path denied")
	// ErrRunLocked is returned when another unattended run holds .mimir/run.lock.
	ErrRunLocked = errors.New("runengine: another run is already active (.mimir/run.lock)")
	// ErrStepLimitExceeded is returned when an unattended run exceeds Limits.MaxSteps.
	ErrStepLimitExceeded = errors.New("runengine: max step limit exceeded")
	// ErrTokenLimitExceeded is returned when an unattended run exceeds Limits.MaxTokens.
	ErrTokenLimitExceeded = errors.New("runengine: max token budget exceeded")
)

// Limits defines the hard resource and verification bounds for an unattended run.
type Limits struct {
	MaxSteps               int           `json:"max_steps"`
	MaxTokens              int           `json:"max_tokens"`
	Timeout                time.Duration `json:"timeout"`
	StaleLockAfter         time.Duration `json:"stale_lock_after"`
	RequireFullSuiteAtGate bool          `json:"require_full_suite_at_gate"`
	RequireEpistemicPass   bool          `json:"require_epistemic_pass"`
}

// DefaultLimits returns production-safe defaults for unattended execution.
func DefaultLimits() Limits {
	return Limits{
		MaxSteps:               16,
		MaxTokens:              120000,
		Timeout:                10 * time.Minute,
		StaleLockAfter:         30 * time.Minute,
		RequireFullSuiteAtGate: true,
		RequireEpistemicPass:   true,
	}
}

// Jail enforces a shell-less workspace boundary rooted at RootDir.
type Jail struct {
	root string
}

// NewJail creates a workspace file jail rooted at rootDir.
func NewJail(rootDir string) (*Jail, error) {
	if strings.TrimSpace(rootDir) == "" {
		rootDir = "."
	}
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return &Jail{root: filepath.Clean(abs)}, nil
}

// Root returns the canonical workspace root directory.
func (j *Jail) Root() string {
	return j.root
}

// Resolve validates that targetPath (relative or absolute) stays strictly inside
// the workspace root, resolving any existing parent symlinks.
func (j *Jail) Resolve(targetPath string) (string, error) {
	if strings.TrimSpace(targetPath) == "" {
		return "", fmt.Errorf("%w: empty path", ErrJailEscape)
	}
	var candidate string
	if filepath.IsAbs(targetPath) {
		candidate = filepath.Clean(targetPath)
	} else {
		candidate = filepath.Clean(filepath.Join(j.root, targetPath))
	}

	// Resolve symlinks on the deepest existing ancestor to catch symlink escapes.
	checkPath := candidate
	for {
		if resolved, err := filepath.EvalSymlinks(checkPath); err == nil {
			suffix, relErr := filepath.Rel(checkPath, candidate)
			if relErr == nil && suffix != "." {
				candidate = filepath.Clean(filepath.Join(resolved, suffix))
			} else {
				candidate = filepath.Clean(resolved)
			}
			break
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			break
		}
		checkPath = parent
	}

	rel, err := filepath.Rel(j.root, candidate)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrJailEscape, err)
	}
	relSlash := filepath.ToSlash(rel)
	if relSlash == ".." || strings.HasPrefix(relSlash, "../") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: %s resolves outside %s", ErrJailEscape, targetPath, j.root)
	}
	return candidate, nil
}

// CheckWrite validates that targetPath is inside the jail and does not mutate
// protected VCS metadata (.git/) or the active run lock (.mimir/run.lock).
func (j *Jail) CheckWrite(targetPath string) (string, error) {
	resolved, err := j.Resolve(targetPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(j.root, resolved)
	if err != nil {
		return "", err
	}
	relSlash := strings.ToLower(filepath.ToSlash(rel))
	if relSlash == ".git" || strings.HasPrefix(relSlash, ".git/") {
		return "", fmt.Errorf("%w: %s", ErrProtectedPath, relSlash)
	}
	if relSlash == ".mimir/run.lock" {
		return "", fmt.Errorf("%w: %s", ErrProtectedPath, relSlash)
	}
	return resolved, nil
}

// ReadFile reads a file strictly inside the workspace jail without invoking a shell.
func (j *Jail) ReadFile(targetPath string) ([]byte, error) {
	resolved, err := j.Resolve(targetPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolved)
}

// WriteFile writes content strictly inside the workspace jail without invoking a shell.
func (j *Jail) WriteFile(targetPath string, data []byte, perm os.FileMode) error {
	resolved, err := j.CheckWrite(targetPath)
	if err != nil {
		return err
	}
	if perm == 0 {
		perm = 0o644
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return err
	}
	return os.WriteFile(resolved, data, perm)
}

// ListDir lists entries in a directory strictly inside the workspace jail.
func (j *Jail) ListDir(targetPath string) ([]string, error) {
	if strings.TrimSpace(targetPath) == "" {
		targetPath = "."
	}
	resolved, err := j.Resolve(targetPath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		out = append(out, name)
	}
	return out, nil
}

// LockMetadata is serialized into .mimir/run.lock while a run is active.
type LockMetadata struct {
	PID       int       `json:"pid"`
	Task      string    `json:"task"`
	StartedAt time.Time `json:"started_at"`
}

// RunLock represents an acquired single-run lockfile (.mimir/run.lock).
type RunLock struct {
	path string
	meta LockMetadata
}

// AcquireLock atomically creates .mimir/run.lock inside rootDir. If an existing
// lockfile is older than staleAfter (and staleAfter > 0), it is treated as stale
// and replaced; otherwise ErrRunLocked is returned.
func AcquireLock(rootDir, task string, staleAfter time.Duration) (*RunLock, error) {
	lockDir := filepath.Join(rootDir, ".mimir")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(lockDir, "run.lock")

	tryCreate := func() (*RunLock, error) {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		meta := LockMetadata{
			PID:       os.Getpid(),
			Task:      task,
			StartedAt: time.Now().UTC(),
		}
		raw, _ := json.MarshalIndent(meta, "", "  ")
		if _, err := f.Write(raw); err != nil {
			_ = os.Remove(lockPath)
			return nil, err
		}
		return &RunLock{path: lockPath, meta: meta}, nil
	}

	lock, err := tryCreate()
	if err == nil {
		return lock, nil
	}
	if !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	// Inspect existing lock for staleness.
	if staleAfter > 0 {
		if raw, readErr := os.ReadFile(lockPath); readErr == nil {
			var existing LockMetadata
			if jsonErr := json.Unmarshal(raw, &existing); jsonErr == nil && !existing.StartedAt.IsZero() {
				if time.Since(existing.StartedAt) > staleAfter {
					_ = os.Remove(lockPath)
					if retryLock, retryErr := tryCreate(); retryErr == nil {
						return retryLock, nil
					}
				}
			}
		} else if fi, statErr := os.Stat(lockPath); statErr == nil {
			if time.Since(fi.ModTime()) > staleAfter {
				_ = os.Remove(lockPath)
				if retryLock, retryErr := tryCreate(); retryErr == nil {
					return retryLock, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("%w at %s", ErrRunLocked, lockPath)
}

// Release removes the .mimir/run.lock file.
func (l *RunLock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}
	err := os.Remove(l.path)
	l.path = ""
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// StepResult captures the outcome of a single unattended step via the Executor seam.
type StepResult struct {
	Reply       string   `json:"reply"`
	TokensUsed  int      `json:"tokens_used"`
	EditedFiles []string `json:"edited_files,omitempty"`
	Done        bool     `json:"done"`
}

// TestRunResult captures the outcome of either Tier-1 targeted tests or Tier-2 full-suite tests.
type TestRunResult struct {
	Tier    string        `json:"tier"` // "targeted_inner_loop" or "full_suite_gate"
	Passed  bool          `json:"passed"`
	Targets []string      `json:"targets"`
	Output  string        `json:"output"`
	Elapsed time.Duration `json:"elapsed"`
}

// Executor is the seam between the run engine and actual LLM/test execution.
// Implementations never invoke an interpolated shell; process execution uses discrete argv.
type Executor interface {
	RunStep(ctx context.Context, prompt string) (StepResult, error)
	RunTargetedTests(ctx context.Context, targets []string) (TestRunResult, error)
	RunFullSuite(ctx context.Context) (TestRunResult, error)
}

// GoTestExecutor implements the test-running side of Executor using direct `go test` argv calls
// (no cmd.exe / powershell / bash shell interpolation) plus an optional step callback.
type GoTestExecutor struct {
	WorkDir string
	StepFn  func(ctx context.Context, prompt string) (StepResult, error)
}

// RunStep delegates to StepFn when configured.
func (g *GoTestExecutor) RunStep(ctx context.Context, prompt string) (StepResult, error) {
	if g.StepFn == nil {
		return StepResult{Reply: "no step runner configured", Done: true}, nil
	}
	return g.StepFn(ctx, prompt)
}

// RunTargetedTests executes `go test` on specific packages or test files (Tier 1 Inner Loop).
func (g *GoTestExecutor) RunTargetedTests(ctx context.Context, targets []string) (TestRunResult, error) {
	start := time.Now()
	pkgs := normalizeGoTargets(g.WorkDir, targets)
	args := append([]string{"test", "-count=1"}, pkgs...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = g.WorkDir
	out, err := cmd.CombinedOutput()
	return TestRunResult{
		Tier:    "targeted_inner_loop",
		Passed:  err == nil,
		Targets: pkgs,
		Output:  strings.TrimSpace(string(out)),
		Elapsed: time.Since(start),
	}, err
}

// RunFullSuite executes `go test -count=1 ./...` across the entire repository (Tier 2 Outer Gate).
func (g *GoTestExecutor) RunFullSuite(ctx context.Context) (TestRunResult, error) {
	start := time.Now()
	args := []string{"test", "-count=1", "./..."}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = g.WorkDir
	out, err := cmd.CombinedOutput()
	return TestRunResult{
		Tier:    "full_suite_gate",
		Passed:  err == nil,
		Targets: []string{"./..."},
		Output:  strings.TrimSpace(string(out)),
		Elapsed: time.Since(start),
	}, err
}

func normalizeGoTargets(workDir string, targets []string) []string {
	if len(targets) == 0 {
		return []string{"./..."}
	}
	seen := make(map[string]bool)
	var pkgs []string
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "./") && !strings.HasSuffix(t, ".go") {
			if !seen[t] {
				seen[t] = true
				pkgs = append(pkgs, t)
			}
			continue
		}
		// Convert file path like "internal/sdlc/sdlc_test.go" to package "./internal/sdlc"
		dir := filepath.ToSlash(filepath.Dir(t))
		if filepath.IsAbs(t) && workDir != "" {
			if rel, err := filepath.Rel(workDir, filepath.Dir(t)); err == nil {
				dir = filepath.ToSlash(rel)
			}
		}
		dir = strings.TrimPrefix(dir, "./")
		pkg := "./" + dir
		if dir == "." || dir == "" {
			pkg = "."
		}
		if !seen[pkg] {
			seen[pkg] = true
			pkgs = append(pkgs, pkg)
		}
	}
	if len(pkgs) == 0 {
		return []string{"./..."}
	}
	return pkgs
}

// GateReport summarizes the deterministic Checklist Gate, Two-Tier Test Gate,
// SDLC Deliverable Gate, and Epistemic Critic Gate for `mimir review` and `mimir ship`.
type GateReport struct {
	Mode              string               `json:"mode"` // "review" or "ship"
	Passed            bool                 `json:"passed"`
	ChecklistPassed   bool                 `json:"checklist_passed"`
	PendingTodos      int                  `json:"pending_todos"`
	SDLCPassed        bool                 `json:"sdlc_passed"`
	SDLCPhase         string               `json:"sdlc_phase"`
	MissingArtifacts  []string             `json:"missing_artifacts,omitempty"`
	FullSuiteResult   TestRunResult        `json:"full_suite_result"`
	EpistemicPassed   bool                 `json:"epistemic_passed"`
	EpistemicReport   *critic.AuditReport  `json:"epistemic_report,omitempty"`
	BlockingReasons   []string             `json:"blocking_reasons,omitempty"`
}

// FormatMarkdown renders a human- and CI-readable summary of the Review/Ship Gate.
func (r GateReport) FormatMarkdown() string {
	var b strings.Builder
	verdict := "PASS"
	if !r.Passed {
		verdict = "FAIL (BLOCKED)"
	}
	b.WriteString(fmt.Sprintf("## Mimir Deterministic %s Gate: `%s`\n", strings.ToUpper(r.Mode), verdict))
	b.WriteString(fmt.Sprintf("- **Checklist Gate**: `pass=%v` (`pending_todos=%d`)\n", r.ChecklistPassed, r.PendingTodos))
	b.WriteString(fmt.Sprintf("- **SDLC Deliverable Gate**: `pass=%v` (`phase=%s`, `missing=%d`)\n", r.SDLCPassed, r.SDLCPhase, len(r.MissingArtifacts)))
	b.WriteString(fmt.Sprintf("- **Tier-2 Full Test Suite (`./...`)**: `pass=%v` (`elapsed=%s`)\n",
		r.FullSuiteResult.Passed, r.FullSuiteResult.Elapsed.Round(time.Millisecond)))
	b.WriteString(fmt.Sprintf("- **Epistemic Critic Gate**: `pass=%v`\n", r.EpistemicPassed))
	if len(r.BlockingReasons) > 0 {
		b.WriteString("\n### Blocking Reasons\n")
		for _, reason := range r.BlockingReasons {
			b.WriteString("- " + reason + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// Engine coordinates unattended runs, the shell-less file jail, the single-run lock,
// and the Two-Tier Testing Law across inner edit loops and outer review/ship gates.
type Engine struct {
	Jail     *Jail
	Limits   Limits
	Executor Executor
	Decider  *decision.Decider
	SDLC     *sdlc.Manager
	Todos    *tools.TodoStore
}

// New creates a new Engine rooted at workDir.
func New(workDir string, exec Executor, limits Limits) (*Engine, error) {
	jail, err := NewJail(workDir)
	if err != nil {
		return nil, err
	}
	if limits.MaxSteps <= 0 {
		limits = DefaultLimits()
	}
	if exec == nil {
		exec = &GoTestExecutor{WorkDir: jail.Root()}
	}
	return &Engine{
		Jail:     jail,
		Limits:   limits,
		Executor: exec,
		Decider:  decision.New("", "", ""),
		SDLC:     sdlc.NewManager(jail.Root()),
		Todos:    tools.NewTodoStore(),
	}, nil
}

// RunInnerLoopTests implements Tier 1 of Mimir's Two-Tier Testing Law:
// During SDLC Steps 4-5 (Implementation & Inner Testing), Huginn-Decider selects
// the affected test files from the Yggdrasil Mind Map so inner edit loops get fast feedback.
func (e *Engine) RunInnerLoopTests(ctx context.Context, editedFiles, candidateTestFiles []string, maxTests int) (TestRunResult, error) {
	if maxTests <= 0 {
		maxTests = 3
	}
	dec := e.Decider
	if dec == nil {
		dec = decision.New("", "", "")
	}
	selected := dec.SelectAffectedTests(ctx, editedFiles, candidateTestFiles, maxTests)
	if len(selected) == 0 {
		selected = editedFiles
	}
	return e.Executor.RunTargetedTests(ctx, selected)
}

// GateOptions provides optional draft/context inputs for the Review and Ship gates.
type GateOptions struct {
	DraftReport      string
	RetrievedContext string
	RequiredPhases   []sdlc.PhaseID
}

// ReviewGate runs the deterministic pre-ship review gate (Tier 2 of the Two-Tier Testing Law):
//  1. Checklist Gate: zero pending/in_progress/blocked todo items.
//  2. SDLC Gate: required phase deliverables exist on disk if specified (or if Waterfall mode is active).
//  3. Full Suite Gate: runs `Executor.RunFullSuite` (`go test ./...`), never just affected tests.
//  4. Epistemic Critic Gate: audits DraftReport against RetrievedContext when provided.
func (e *Engine) ReviewGate(ctx context.Context, opts GateOptions) GateReport {
	return e.evaluateGate(ctx, "review", opts)
}

// ShipGate runs the mandatory production release gate (`mimir ship`).
// In addition to all ReviewGate checks, if passed it signs off SDLC Step 5 (Testing)
// and advances the SDLC state machine to Step 6 (Deployment).
func (e *Engine) ShipGate(ctx context.Context, opts GateOptions) GateReport {
	rep := e.evaluateGate(ctx, "ship", opts)
	if rep.Passed && e.SDLC != nil {
		e.SDLC.SignOff(sdlc.PhaseTesting)
		_, _ = e.SDLC.SetPhase(string(sdlc.PhaseDeployment))
		rep.SDLCPhase = string(e.SDLC.CurrentPhase().ID)
	}
	return rep
}

func (e *Engine) evaluateGate(ctx context.Context, mode string, opts GateOptions) GateReport {
	rep := GateReport{
		Mode:            mode,
		ChecklistPassed: true,
		SDLCPassed:      true,
		EpistemicPassed: true,
	}

	// 1. Checklist Gate
	if e.Todos != nil {
		for _, item := range e.Todos.List() {
			if item.Status != tools.TodoCompleted {
				rep.PendingTodos++
			}
		}
		if rep.PendingTodos > 0 {
			rep.ChecklistPassed = false
			rep.BlockingReasons = append(rep.BlockingReasons,
				fmt.Sprintf("checklist gate failed: %d todo item(s) are not marked completed", rep.PendingTodos))
		}
	}

	// 2. SDLC Deliverable Gate
	if e.SDLC != nil {
		cur := e.SDLC.CurrentPhase()
		rep.SDLCPhase = string(cur.ID)
		required := opts.RequiredPhases
		if len(required) == 0 && e.SDLC.Methodology() == sdlc.MethodWaterfall {
			// In Waterfall mode, all phases up to the current step must have deliverables on disk.
			for _, p := range sdlc.AllPhases() {
				if p.Step <= cur.Step {
					required = append(required, p.ID)
				}
			}
		}
		for _, pid := range required {
			if spec, ok := sdlc.LookupPhase(string(pid)); ok {
				full := filepath.Join(e.Jail.Root(), filepath.FromSlash(spec.ArtifactFile))
				if _, err := os.Stat(full); err != nil {
					rep.MissingArtifacts = append(rep.MissingArtifacts, spec.ArtifactFile)
				}
			}
		}
		if len(rep.MissingArtifacts) > 0 {
			rep.SDLCPassed = false
			rep.BlockingReasons = append(rep.BlockingReasons,
				"missing required SDLC deliverable(s): "+strings.Join(rep.MissingArtifacts, ", "))
		}
	}

	// 3. Tier-2 Full Test Suite Gate (Law: always run full suite at the gate)
	if e.Limits.RequireFullSuiteAtGate {
		res, err := e.Executor.RunFullSuite(ctx)
		rep.FullSuiteResult = res
		if err != nil || !res.Passed {
			rep.FullSuiteResult.Passed = false
			rep.BlockingReasons = append(rep.BlockingReasons,
				"full test suite failed at outer gate (`go test ./...` must be 100% green)")
		}
	} else {
		rep.FullSuiteResult = TestRunResult{Tier: "full_suite_gate", Passed: true}
	}

	// 4. Epistemic Critic Gate
	if e.Limits.RequireEpistemicPass && strings.TrimSpace(opts.DraftReport) != "" {
		audit := critic.AuditDeterministic(opts.RetrievedContext, opts.DraftReport)
		rep.EpistemicReport = &audit
		if !audit.PassAudit {
			rep.EpistemicPassed = false
			rep.BlockingReasons = append(rep.BlockingReasons,
				"epistemic critic rejected draft report ("+audit.RoutingReason+")")
		}
	}

	rep.Passed = rep.ChecklistPassed && rep.SDLCPassed && rep.FullSuiteResult.Passed && rep.EpistemicPassed
	return rep
}

// UnattendedRunSummary records the trace and gate verdict of an unattended run.
type UnattendedRunSummary struct {
	Task        string            `json:"task"`
	Steps       int               `json:"steps"`
	TokensUsed  int               `json:"tokens_used"`
	InnerTests  []TestRunResult   `json:"inner_tests,omitempty"`
	FinalReply  string            `json:"final_reply"`
	Gate        GateReport        `json:"gate"`
}

// RunUnattended executes an unattended task inside the single-run lockfile (.mimir/run.lock),
// enforcing Limits (MaxSteps, MaxTokens, Timeout), running Tier-1 targeted tests whenever
// files are edited, and enforcing the Tier-2 ReviewGate before finishing.
func (e *Engine) RunUnattended(ctx context.Context, task string, candidateTestFiles []string) (*UnattendedRunSummary, error) {
	lock, err := AcquireLock(e.Jail.Root(), task, e.Limits.StaleLockAfter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = lock.Release() }()

	if e.Limits.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.Limits.Timeout)
		defer cancel()
	}

	summary := &UnattendedRunSummary{Task: task}
	prompt := task

	for step := 1; step <= e.Limits.MaxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		res, err := e.Executor.RunStep(ctx, prompt)
		if err != nil {
			return summary, err
		}
		summary.Steps = step
		summary.TokensUsed += res.TokensUsed
		summary.FinalReply = res.Reply

		if e.Limits.MaxTokens > 0 && summary.TokensUsed > e.Limits.MaxTokens {
			return summary, fmt.Errorf("%w: used %d > limit %d", ErrTokenLimitExceeded, summary.TokensUsed, e.Limits.MaxTokens)
		}

		// Verify all edited files remain inside the workspace jail.
		for _, edited := range res.EditedFiles {
			if _, jailErr := e.Jail.CheckWrite(edited); jailErr != nil {
				return summary, jailErr
			}
		}

		// Tier 1 (Inner Loop): run targeted affected tests when files were modified.
		if len(res.EditedFiles) > 0 {
			tRes, _ := e.RunInnerLoopTests(ctx, res.EditedFiles, candidateTestFiles, 3)
			summary.InnerTests = append(summary.InnerTests, tRes)
			if !tRes.Passed && !res.Done {
				prompt = "Tier-1 targeted tests failed:\n" + tRes.Output + "\nFix the failing tests before continuing."
				continue
			}
		}

		if res.Done {
			summary.Gate = e.ReviewGate(ctx, GateOptions{})
			return summary, nil
		}
	}

	return summary, fmt.Errorf("%w: exceeded %d steps", ErrStepLimitExceeded, e.Limits.MaxSteps)
}
