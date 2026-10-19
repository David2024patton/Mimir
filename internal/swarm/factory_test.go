package swarm

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
)

func TestFactory_SubmitAndListTasks(t *testing.T) {
	dir := t.TempDir()
	f := NewFactory(FactoryConfig{
		RootDir:      dir,
		MaxWorkers:   2,
		TargetBranch: "main",
	})

	task1, err := f.Submit("Build Telephony Gateway", "Integrate Twilio voice and SMS webhooks.")
	if err != nil {
		t.Fatalf("Submit task1 failed: %v", err)
	}
	if task1.ID != "task-1" {
		t.Errorf("expected task-1, got %s", task1.ID)
	}
	if task1.State != StateBacklog {
		t.Errorf("expected StateBacklog, got %s", task1.State)
	}
	if !strings.Contains(task1.Branch, "task-1") {
		t.Errorf("expected branch to contain task-1, got %s", task1.Branch)
	}

	task2, err := f.Submit("SEO and LLMs.txt", "Generate robots.txt and llms.txt.")
	if err != nil {
		t.Fatalf("Submit task2 failed: %v", err)
	}
	if task2.ID != "task-2" {
		t.Errorf("expected task-2, got %s", task2.ID)
	}

	tasks := f.ListTasks()
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "Build Telephony Gateway" || tasks[1].Title != "SEO and LLMs.txt" {
		t.Errorf("unexpected task order: %+v", tasks)
	}

	status := f.FormatStatus()
	if !strings.Contains(status, "Build Telephony Gateway") || !strings.Contains(status, "SEO and LLMs.txt") {
		t.Errorf("FormatStatus missing task titles:\n%s", status)
	}
}

func TestFactory_StateTransitionsAndArtifacts(t *testing.T) {
	dir := t.TempDir()
	f := NewFactory(FactoryConfig{
		RootDir:      dir,
		TargetBranch: "main",
	})

	task, err := f.Submit("Add Feature X", "Detailed description")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	f.updateState(task.ID, StateSpec, "odin")
	f.recordArtifact(task.ID, ".mimir/sdlc/01-SPEC-task-1.md")

	got, ok := f.GetTask(task.ID)
	if !ok {
		t.Fatalf("GetTask failed")
	}
	if got.State != StateSpec || got.AssignedTo != "odin" {
		t.Errorf("expected StateSpec/odin, got %s/%s", got.State, got.AssignedTo)
	}
	if len(got.Artifacts) != 1 || got.Artifacts[0] != ".mimir/sdlc/01-SPEC-task-1.md" {
		t.Errorf("unexpected artifacts: %v", got.Artifacts)
	}

	// Test failure state transition
	f.failTask(task.ID, "simulated test failure")
	failed, _ := f.GetTask(task.ID)
	if failed.State != StateFailed || !strings.Contains(failed.Error, "simulated test failure") {
		t.Errorf("expected failed state with error message, got %+v", failed)
	}
}

// N12: IDs come from a persisted counter, so a factory reloaded from disk
// never reissues an ID that an earlier instance already handed out.
func TestFactory_UniqueIDsAcrossReload(t *testing.T) {
	dir := t.TempDir()
	cfg := FactoryConfig{RootDir: dir, TargetBranch: "main"}

	f1 := NewFactory(cfg)
	if _, err := f1.Submit("One", "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := f1.Submit("Two", "second"); err != nil {
		t.Fatal(err)
	}

	f2 := NewFactory(cfg) // reload from disk
	third, err := f2.Submit("Three", "third")
	if err != nil {
		t.Fatal(err)
	}
	if third.ID != "task-3" {
		t.Fatalf("expected task-3 after reload, got %s", third.ID)
	}
	seen := map[string]bool{}
	for _, task := range f2.ListTasks() {
		if seen[task.ID] {
			t.Fatalf("duplicate task ID %s", task.ID)
		}
		seen[task.ID] = true
	}
}

// N12: branch slugs are git-ref-safe regardless of title punctuation.
func TestFactory_SanitizesBranchSlug(t *testing.T) {
	dir := t.TempDir()
	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})

	task, err := f.Submit("Fix: Auth!! (ASAP) ### v2.0", "desc")
	if err != nil {
		t.Fatal(err)
	}
	want := "swarm/task-1-fix-auth-asap-v2-0"
	if task.Branch != want {
		t.Errorf("expected branch %q, got %q", want, task.Branch)
	}

	safe := regexp.MustCompile(`^swarm/task-\d+-[a-z0-9-]+$`)
	bang, err := f.Submit("!!!", "desc")
	if err != nil {
		t.Fatal(err)
	}
	if !safe.MatchString(bang.Branch) {
		t.Errorf("branch for punctuation-only title is not ref-safe: %q", bang.Branch)
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	gitCmd(t, dir, "init")
	gitCmd(t, dir, "symbolic-ref", "HEAD", "refs/heads/main")
	gitCmd(t, dir, "config", "user.email", "swarm-test@example.com")
	gitCmd(t, dir, "config", "user.name", "Swarm Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "init")
}

// N6: with no agent configured, ExecuteTask must stop honestly in
// StateBlocked instead of fabricating an implementation.
func TestExecuteTask_BlockedWithoutAgent(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})
	task, err := f.Submit("Add Feature Y", "Detailed description")
	if err != nil {
		t.Fatal(err)
	}

	err = f.ExecuteTask(context.Background(), task.ID)
	if !errors.Is(err, ErrNoAgentConfigured) {
		t.Fatalf("expected ErrNoAgentConfigured, got %v", err)
	}
	got, _ := f.GetTask(task.ID)
	if got.State != StateBlocked {
		t.Fatalf("expected StateBlocked, got %s", got.State)
	}
}

// N6: ExecuteTask never merges on its own. A gate-passed task waits in
// StateAwaitingMerge until ApproveMerge is called, and ApproveMerge lands
// the merge on TargetBranch only.
func TestApproveMerge_GatesAndTargetsBranch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})
	task, err := f.Submit("Add Feature Z", "Detailed description")
	if err != nil {
		t.Fatal(err)
	}

	// Simulate Thor's committed work on the task branch.
	wtDir, err := f.setupWorktree(task.Branch)
	if err != nil {
		t.Fatalf("setupWorktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wtDir, "feature.txt"), []byte("z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, wtDir, "add", ".")
	gitCmd(t, wtDir, "commit", "-m", "feat: feature z")
	f.teardownWorktree(wtDir)
	f.updateState(task.ID, StateAwaitingMerge, "none")

	// Approving from the wrong checked-out branch is refused and the task
	// stays awaiting merge (retryable).
	gitCmd(t, dir, "checkout", "-b", "other")
	if err := f.ApproveMerge(context.Background(), task.ID); err == nil {
		t.Fatal("expected refusal when main worktree is not on the target branch")
	}
	got, _ := f.GetTask(task.ID)
	if got.State != StateAwaitingMerge {
		t.Fatalf("expected task to stay StateAwaitingMerge after refusal, got %s", got.State)
	}
	gitCmd(t, dir, "checkout", "main")

	// Approval on the target branch merges and flips the state.
	if err := f.ApproveMerge(context.Background(), task.ID); err != nil {
		t.Fatalf("ApproveMerge: %v", err)
	}
	got, _ = f.GetTask(task.ID)
	if got.State != StateMerged {
		t.Fatalf("expected StateMerged, got %s", got.State)
	}
	if _, err := os.Stat(filepath.Join(dir, "feature.txt")); err != nil {
		t.Fatalf("merged file missing from target branch worktree: %v", err)
	}

	// A task that never passed the gates cannot be approved.
	other, err := f.Submit("Backlog Task", "not run")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.ApproveMerge(context.Background(), other.ID); err == nil {
		t.Fatal("expected error approving a task that is not awaiting merge")
	}
}

// FIX 3 (audit): saveToDisk swallowed every error, so Submit reported
// success for tasks that were never persisted (and the persisted nextID
// guarantee silently broke with them). A failed persist must fail the
// submit and roll back the in-memory task.
func TestFactory_SubmitFailsWhenStoreUnwritable(t *testing.T) {
	dir := t.TempDir()
	// Block the store path: .mimir exists as a regular file, so creating
	// .mimir/swarm fails regardless of process privileges.
	if err := os.WriteFile(filepath.Join(dir, ".mimir"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})
	if _, err := f.Submit("Doomed Task", "cannot persist"); err == nil {
		t.Fatal("expected Submit to fail when the task store cannot be written")
	}
	if got := f.ListTasks(); len(got) != 0 {
		t.Fatalf("failed Submit must roll back the in-memory task, got %d tasks", len(got))
	}
}

// FIX 3: a corrupt task store must not be silently discarded. It is
// preserved under a side name and the factory starts empty; later saves
// write a fresh store and never touch the preserved bytes.
func TestFactory_CorruptStorePreserved(t *testing.T) {
	dir := t.TempDir()
	storeDir := filepath.Join(dir, ".mimir", "swarm")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	garbage := []byte("{not json!!")
	storePath := filepath.Join(storeDir, "tasks.json")
	if err := os.WriteFile(storePath, garbage, 0o644); err != nil {
		t.Fatal(err)
	}

	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})
	if got := f.ListTasks(); len(got) != 0 {
		t.Fatalf("corrupt store must start empty, got %d tasks", len(got))
	}
	matches, err := filepath.Glob(storePath + ".corrupt-*")
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected exactly one preserved corrupt store, got %v (err %v)", matches, err)
	}
	preserved, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(preserved) != string(garbage) {
		t.Fatalf("preserved store content changed: %q", preserved)
	}

	// The factory still works, and the fresh save does not touch the
	// preserved copy.
	if _, err := f.Submit("Fresh Task", "after corruption"); err != nil {
		t.Fatalf("Submit after corrupt store recovery failed: %v", err)
	}
	preserved2, _ := os.ReadFile(matches[0])
	if string(preserved2) != string(garbage) {
		t.Fatal("preserved corrupt store was modified by a later save")
	}
	raw, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) == string(garbage) {
		t.Fatal("fresh store was not written after recovery")
	}
}

// stubProvider is a deterministic llm.Provider that finishes immediately
// with no tool calls, so Thor's agent pass completes without a model.
type stubProvider struct{}

func (stubProvider) ID() string { return "stub" }

func (stubProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{Content: "done", FinishReason: "stop"}, nil
}

func (stubProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	return nil, nil
}

// captureSlog redirects the default slog logger into a buffer for the
// duration of the test.
func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

// FIX 5 (audit): the factory's trust boundaries (submit, gate outcome,
// merge approval, merge) must leave a log record carrying only the task
// ID and branch, never task description contents.
func TestFactory_TrustBoundaryLogging(t *testing.T) {
	buf := captureSlog(t)
	dir := t.TempDir()
	initGitRepo(t, dir)
	// Loki runs `go test ./...` in the worktree, so the repo needs a Go
	// module with a buildable package.
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module swarmtest\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", ".")
	gitCmd(t, dir, "commit", "-m", "add go module")

	f := NewFactory(FactoryConfig{
		RootDir:      dir,
		TargetBranch: "main",
		AgentConfig: agent.Config{
			Provider: stubProvider{},
			Cortex:   cortex.NewMemoryStore(),
			Model:    "stub-model",
		},
	})
	task, err := f.Submit("Logging Feature", "TOPSECRET description marker")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.ExecuteTask(context.Background(), task.ID); err != nil {
		t.Fatalf("ExecuteTask: %v", err)
	}
	got, _ := f.GetTask(task.ID)
	if got.State != StateAwaitingMerge {
		t.Fatalf("expected StateAwaitingMerge, got %s (error %q)", got.State, got.Error)
	}
	if err := f.ApproveMerge(context.Background(), task.ID); err != nil {
		t.Fatalf("ApproveMerge: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"task submitted", "passed all gates", "merge approved", "merge performed", task.ID, task.Branch} {
		if !strings.Contains(out, want) {
			t.Errorf("log is missing %q\nlog:\n%s", want, out)
		}
	}
	if strings.Contains(out, "TOPSECRET") {
		t.Fatalf("log leaked task description contents:\n%s", out)
	}
}

// FIX 5: a gate or pipeline failure is logged through the same funnel,
// again with task ID and branch only.
func TestFactory_FailureLogging(t *testing.T) {
	buf := captureSlog(t)
	dir := t.TempDir() // deliberately not a git repo: worktree setup fails
	f := NewFactory(FactoryConfig{RootDir: dir, TargetBranch: "main"})
	task, err := f.Submit("Will Fail", "TOPSECRET failure marker")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.ExecuteTask(context.Background(), task.ID); err == nil {
		t.Fatal("expected ExecuteTask to fail outside a git repo")
	}
	out := buf.String()
	for _, want := range []string{"task failed", task.ID, task.Branch} {
		if !strings.Contains(out, want) {
			t.Errorf("log is missing %q\nlog:\n%s", want, out)
		}
	}
	if strings.Contains(out, "TOPSECRET") {
		t.Fatalf("log leaked task description contents:\n%s", out)
	}
}
