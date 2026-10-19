package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/critic"
	"github.com/David2024patton/Mimir/internal/runengine"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TaskState tracks the lifecycle of an autonomous factory task through the Norse personas.
type TaskState string

const (
	StateBacklog       TaskState = "backlog"        // Waiting for specification
	StateSpec          TaskState = "spec"           // Odin (Orchestrator) writing requirements
	StateImplement     TaskState = "implement"      // Thor (Builder) developing in worktree
	StateTest          TaskState = "test"           // Loki (Tester/Debugger) validating and fixing
	StateReview        TaskState = "review"         // Heimdall (Auditor) running SEO/GEO and web audit
	StateAwaitingMerge TaskState = "awaiting_merge" // All gates passed; a human must call ApproveMerge
	StateMerged        TaskState = "merged"         // Merged back to target branch after human approval
	StateBlocked       TaskState = "blocked"        // Cannot proceed without external input (e.g. no agent configured)
	StateFailed        TaskState = "failed"         // Halted due to unrecoverable test or review failure
)

// ErrNoAgentConfigured is returned when a task reaches the Thor build phase
// but the factory has no agent configured to do the work. The task is left
// in StateBlocked, never reported as built.
var ErrNoAgentConfigured = errors.New("swarm: no agent configured (FactoryConfig.AgentConfig.Provider is nil); Thor cannot implement tasks")

// Task represents a discrete unit of work assigned to the autonomous software factory.
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Branch      string            `json:"branch"`
	WorktreeDir string            `json:"worktree_dir,omitempty"`
	State       TaskState         `json:"state"`
	AssignedTo  string            `json:"assigned_to"` // "odin" | "thor" | "loki" | "heimdall"
	Error       string            `json:"error,omitempty"`
	Artifacts   []string          `json:"artifacts,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FactoryConfig configures the autonomous software factory queue.
type FactoryConfig struct {
	RootDir      string
	MaxWorkers   int
	AgentConfig  agent.Config
	RunLimits    runengine.Limits
	TargetBranch string // default "main"
}

// Factory manages parallel multi-agent worker pipelines operating in isolated git worktrees.
type Factory struct {
	mu           sync.RWMutex
	cfg          FactoryConfig
	tasks        map[string]*Task
	taskOrder    []string
	nextID       int // persisted monotonic counter for task IDs (N12)
	worktreeRoot string
}

// NewFactory creates an autonomous software factory instance.
func NewFactory(cfg FactoryConfig) *Factory {
	if cfg.RootDir == "" {
		cfg.RootDir = "."
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 4
	}
	if cfg.TargetBranch == "" {
		cfg.TargetBranch = "main"
	}
	wtBase := filepath.Join(cfg.RootDir, ".mimir", "worktrees")
	fac := &Factory{
		cfg:          cfg,
		tasks:        make(map[string]*Task),
		taskOrder:    make([]string, 0),
		worktreeRoot: wtBase,
	}
	fac.loadFromDisk()
	return fac
}

func (f *Factory) storagePath() string {
	return filepath.Join(f.cfg.RootDir, ".mimir", "swarm", "tasks.json")
}

// factoryStore is the on-disk form of the factory queue. NextID persists the
// task ID counter so IDs stay unique across process restarts and reloads.
type factoryStore struct {
	Order  []string         `json:"order"`
	Tasks  map[string]*Task `json:"tasks"`
	NextID int              `json:"next_id"`
}

func (f *Factory) loadFromDisk() {
	raw, err := os.ReadFile(f.storagePath())
	if err != nil {
		return
	}
	var data factoryStore
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	f.taskOrder = data.Order
	f.tasks = data.Tasks
	if f.tasks == nil {
		f.tasks = make(map[string]*Task)
	}
	// Legacy stores have no counter: derive it from the highest existing ID
	// so a reloaded factory never reissues an ID (N12).
	f.nextID = data.NextID
	for id := range f.tasks {
		if n, err := strconv.Atoi(strings.TrimPrefix(id, "task-")); err == nil && n > f.nextID {
			f.nextID = n
		}
	}
}

func (f *Factory) saveToDisk() {
	data := factoryStore{
		Order:  f.taskOrder,
		Tasks:  f.tasks,
		NextID: f.nextID,
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}
	p := f.storagePath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, append(raw, '\n'), 0o644)
}

// Submit enqueues a new feature or bugfix task into the factory backlog.
func (f *Factory) Submit(title, description string) (*Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("factory: task title cannot be empty")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// IDs come from a persisted monotonic counter, never from the live task
	// count, so reloads and deletions cannot reissue an ID (N12).
	f.nextID++
	id := fmt.Sprintf("task-%d", f.nextID)
	branch := fmt.Sprintf("swarm/%s-%s", id, sanitizeSlug(title))

	t := &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Branch:      branch,
		State:       StateBacklog,
		AssignedTo:  "odin",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Metadata:    make(map[string]string),
	}

	f.tasks[id] = t
	f.taskOrder = append(f.taskOrder, id)
	f.saveToDisk()
	return t, nil
}

// sanitizeSlug reduces a task title to a git-ref-safe branch slug:
// lowercase [a-z0-9] words joined by single dashes, at most 30 chars,
// with a fallback when nothing usable remains (N12).
func sanitizeSlug(title string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 30 {
		slug = strings.TrimRight(slug[:30], "-")
	}
	if slug == "" {
		slug = "task"
	}
	return slug
}

// ListTasks returns all tasks in submission order.
func (f *Factory) ListTasks() []*Task {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make([]*Task, 0, len(f.taskOrder))
	for _, id := range f.taskOrder {
		if t, ok := f.tasks[id]; ok {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out
}

// GetTask returns a task by ID.
func (f *Factory) GetTask(id string) (*Task, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	t, ok := f.tasks[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

// ExecuteTask runs a single task through the full Odin -> Thor -> Loki -> Heimdall pipeline.
func (f *Factory) ExecuteTask(ctx context.Context, taskID string) error {
	f.mu.Lock()
	t, ok := f.tasks[taskID]
	if !ok {
		f.mu.Unlock()
		return fmt.Errorf("factory: task %q not found", taskID)
	}
	f.mu.Unlock()

	// 1. Odin: Specification Phase
	f.updateState(taskID, StateSpec, "odin")
	specDoc, err := f.runOdinSpec(ctx, t)
	if err != nil {
		f.failTask(taskID, fmt.Sprintf("Odin spec phase failed: %v", err))
		return err
	}
	f.recordArtifact(taskID, specDoc)

	// 2. Thor: Implementation in isolated Git Worktree
	f.updateState(taskID, StateImplement, "thor")
	wtDir, err := f.setupWorktree(t.Branch)
	if err != nil {
		f.failTask(taskID, fmt.Sprintf("Thor worktree setup failed: %v", err))
		return err
	}
	defer f.teardownWorktree(wtDir)

	f.mu.Lock()
	t.WorktreeDir = wtDir
	f.mu.Unlock()

	if err := f.runThorImplement(ctx, t, wtDir, specDoc); err != nil {
		if errors.Is(err, ErrNoAgentConfigured) {
			// Honest stop, not a failure and never a fake build: the task
			// waits in StateBlocked until the factory is given an agent.
			f.blockTask(taskID, err.Error())
			return err
		}
		f.failTask(taskID, fmt.Sprintf("Thor implementation failed: %v", err))
		return err
	}

	// 3. Loki: Testing and Adversarial Debugging
	f.updateState(taskID, StateTest, "loki")
	if err := f.runLokiTesting(ctx, t, wtDir); err != nil {
		f.failTask(taskID, fmt.Sprintf("Loki test phase failed: %v", err))
		return err
	}

	// 4. Heimdall: Review Gate, Epistemic Critic & Web Audit
	f.updateState(taskID, StateReview, "heimdall")
	if err := f.runHeimdallReview(ctx, t, wtDir); err != nil {
		f.failTask(taskID, fmt.Sprintf("Heimdall review gate failed: %v", err))
		return err
	}

	// 5. Stop for human merge approval (N6). Passing every gate does NOT
	// merge: the task waits in StateAwaitingMerge until a human calls
	// ApproveMerge from the CLI or TUI. The committed task branch is left
	// intact; only the worktree is torn down.
	f.updateState(taskID, StateAwaitingMerge, "none")
	return nil
}

// ApproveMerge is the human approval step for a task that passed all
// factory gates. It merges the task branch into the factory TargetBranch
// and marks the task StateMerged. The merge is refused when the main
// worktree is not currently on TargetBranch, so an approval can never land
// code on whatever branch happens to be checked out (N6).
func (f *Factory) ApproveMerge(ctx context.Context, taskID string) error {
	f.mu.RLock()
	t, ok := f.tasks[taskID]
	f.mu.RUnlock()
	if !ok {
		return fmt.Errorf("factory: task %q not found", taskID)
	}
	if t.State != StateAwaitingMerge {
		return fmt.Errorf("factory: task %q is in state %q, not %q; only gate-passed tasks can be approved for merge", taskID, t.State, StateAwaitingMerge)
	}

	// A refusal (wrong branch checked out) leaves the task awaiting merge so
	// the approval can be retried; only a failed merge attempt fails it.
	if current, err := f.currentBranch(ctx); err != nil {
		return err
	} else if current != f.cfg.TargetBranch {
		return fmt.Errorf("factory: refusing to merge %s: main worktree is on branch %q but the target branch is %q; check out %q first", t.Branch, current, f.cfg.TargetBranch, f.cfg.TargetBranch)
	}

	if err := f.mergeTaskBranch(ctx, t.Branch); err != nil {
		f.failTask(taskID, fmt.Sprintf("Git merge failed: %v", err))
		return err
	}

	f.updateState(taskID, StateMerged, "none")
	return nil
}

// currentBranch returns the branch checked out in the main worktree.
func (f *Factory) currentBranch(ctx context.Context) (string, error) {
	head := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	head.Dir = f.cfg.RootDir
	out, err := head.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot determine current branch of main worktree: %s (%v)", string(out), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (f *Factory) updateState(id string, st TaskState, assignedTo string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.tasks[id]; ok {
		t.State = st
		t.AssignedTo = assignedTo
		t.UpdatedAt = time.Now().UTC()
		f.saveToDisk()
	}
}

func (f *Factory) failTask(id string, errMsg string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.tasks[id]; ok {
		t.State = StateFailed
		t.Error = errMsg
		t.UpdatedAt = time.Now().UTC()
		f.saveToDisk()
	}
}

func (f *Factory) blockTask(id string, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.tasks[id]; ok {
		t.State = StateBlocked
		t.Error = reason
		t.UpdatedAt = time.Now().UTC()
		f.saveToDisk()
	}
}

func (f *Factory) recordArtifact(id, path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.tasks[id]; ok {
		t.Artifacts = append(t.Artifacts, path)
		f.saveToDisk()
	}
}

func (f *Factory) runOdinSpec(ctx context.Context, t *Task) (string, error) {
	specPath := filepath.Join(f.cfg.RootDir, ".mimir", "sdlc", fmt.Sprintf("01-SPEC-%s.md", t.ID))
	_ = os.MkdirAll(filepath.Dir(specPath), 0o755)

	content := fmt.Sprintf("# Factory Task Specification: %s (%s)\n\n"+
		"- **Objective**: %s\n"+
		"- **Branch**: `%s`\n"+
		"- **Assigned**: Odin (Orchestrator)\n\n"+
		"## Functional Scope\n%s\n\n"+
		"## Verification Law\n"+
		"- All unit and integration tests must pass cleanly.\n"+
		"- Zero em dashes allowed in code, comments, or documentation.\n"+
		"- Strict anti-slop copy guidelines enforced.\n",
		t.Title, t.ID, t.Title, t.Branch, t.Description)

	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return specPath, nil
}

func (f *Factory) setupWorktree(branch string) (string, error) {
	_ = os.MkdirAll(f.worktreeRoot, 0o755)
	cleanName := strings.ReplaceAll(branch, "/", "-")
	wtDir := filepath.Join(f.worktreeRoot, cleanName)

	if _, err := os.Stat(wtDir); err == nil {
		_ = exec.Command("git", "worktree", "remove", "--force", wtDir).Run()
	}

	// Create worktree with a new branch branched off TargetBranch
	cmd := exec.Command("git", "worktree", "add", "-b", branch, wtDir, f.cfg.TargetBranch)
	cmd.Dir = f.cfg.RootDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// If branch already exists, checkout without -b
		fallback := exec.Command("git", "worktree", "add", wtDir, branch)
		fallback.Dir = f.cfg.RootDir
		if fOut, fErr := fallback.CombinedOutput(); fErr != nil {
			return "", fmt.Errorf("git worktree add failed: %s (%s)", string(out), string(fOut))
		}
	}

	return wtDir, nil
}

func (f *Factory) teardownWorktree(wtDir string) {
	if wtDir == "" {
		return
	}
	cmd := exec.Command("git", "worktree", "remove", "--force", wtDir)
	cmd.Dir = f.cfg.RootDir
	_ = cmd.Run()
}

// runThorImplement runs one real agent pass inside the task worktree using
// the factory's configured agent, then commits whatever the agent produced
// to the task branch so the branch (not the disposable worktree) carries
// the work into the review and merge gates. Without a configured agent it
// returns ErrNoAgentConfigured: Thor never fabricates a build record and
// claims work it did not do.
func (f *Factory) runThorImplement(ctx context.Context, t *Task, wtDir, specDoc string) error {
	if f.cfg.AgentConfig.Provider == nil {
		return ErrNoAgentConfigured
	}

	cfg := f.cfg.AgentConfig
	// The worker's tools must be rooted in the task worktree, not the main
	// checkout, so every file and shell action stays inside the jail.
	cfg.Tools = tools.Default(wtDir)
	if cfg.Mode == "" {
		// Autonomous worker: there is no human at this terminal to answer
		// approval prompts. The policy gate (cfg.Gate) still applies.
		cfg.Mode = agent.ModeBypass
	}
	worker := agent.New(cfg)

	specBody := ""
	if raw, err := os.ReadFile(specDoc); err == nil {
		specBody = string(raw)
	}
	prompt := fmt.Sprintf("You are Thor, the builder in an autonomous software factory. "+
		"Implement the task below inside the current workspace, which is a dedicated git worktree on branch %s. "+
		"Make the smallest complete change that satisfies the spec, run the project's tests if practical, then stop.\n\n"+
		"Task: %s\n\nDescription: %s\n\nSpecification:\n%s",
		t.Branch, t.Title, t.Description, specBody)
	if _, err := worker.Run(ctx, prompt, nil); err != nil {
		return fmt.Errorf("thor agent run failed: %w", err)
	}

	buildRecord := filepath.Join(wtDir, ".mimir", "sdlc", fmt.Sprintf("04-BUILD-%s.md", t.ID))
	_ = os.MkdirAll(filepath.Dir(buildRecord), 0o755)
	content := fmt.Sprintf("# Thor Implementation Record: %s\n\n"+
		"- Task ID: `%s`\n"+
		"- Worktree: `%s`\n"+
		"- Status: Agent pass complete; changes committed to `%s` for Loki verification\n",
		t.Title, t.ID, wtDir, t.Branch)
	if err := os.WriteFile(buildRecord, []byte(content), 0o644); err != nil {
		return err
	}

	return f.commitWorktreeChanges(ctx, wtDir, t)
}

// commitWorktreeChanges commits all pending worktree changes to the task
// branch. A clean worktree (agent changed nothing) is not an error.
func (f *Factory) commitWorktreeChanges(ctx context.Context, wtDir string, t *Task) error {
	status := exec.CommandContext(ctx, "git", "status", "--porcelain")
	status.Dir = wtDir
	out, err := status.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status in worktree failed: %s (%v)", string(out), err)
	}
	if strings.TrimSpace(string(out)) == "" {
		return nil
	}
	add := exec.CommandContext(ctx, "git", "add", "-A")
	add.Dir = wtDir
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git add in worktree failed: %s (%v)", string(out), err)
	}
	commit := exec.CommandContext(ctx, "git", "commit", "-m", fmt.Sprintf("feat(swarm): %s (%s)", t.Title, t.ID))
	commit.Dir = wtDir
	if out, err := commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit in worktree failed: %s (%v)", string(out), err)
	}
	return nil
}

func (f *Factory) runLokiTesting(ctx context.Context, t *Task, wtDir string) error {
	// Execute Loki test verification inside the worktree
	cmd := exec.CommandContext(ctx, "go", "test", "-count=1", "./...")
	cmd.Dir = wtDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tests failed in worktree:\n%s", string(out))
	}
	return nil
}

func (f *Factory) runHeimdallReview(ctx context.Context, t *Task, wtDir string) error {
	// Heimdall runs the web audit and epistemic critic checks on modified files
	rep, err := tools.AuditWebDirectory(wtDir)
	if err == nil && !rep.PassAudit {
		return fmt.Errorf("web audit detector failed with %d findings", len(rep.Findings))
	}

	criticCheck := critic.AuditDeterministic("", t.Description)
	if !criticCheck.PassAudit {
		return fmt.Errorf("epistemic critic flagged task description: %s", criticCheck.RoutingReason)
	}

	return nil
}

func (f *Factory) mergeTaskBranch(ctx context.Context, branch string) error {
	// The merge must land on the configured target branch and nowhere else
	// (N6): refuse when the main worktree has a different branch checked out.
	current, err := f.currentBranch(ctx)
	if err != nil {
		return err
	}
	if current != f.cfg.TargetBranch {
		return fmt.Errorf("refusing to merge %s: main worktree is on branch %q but the target branch is %q; check out %q first", branch, current, f.cfg.TargetBranch, f.cfg.TargetBranch)
	}

	cmd := exec.CommandContext(ctx, "git", "merge", "--no-ff", "-m", fmt.Sprintf("feat(swarm): merge %s", branch), branch)
	cmd.Dir = f.cfg.RootDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// Abort conflicted merge if needed
		abort := exec.CommandContext(ctx, "git", "merge", "--abort")
		abort.Dir = f.cfg.RootDir
		_ = abort.Run()
		return fmt.Errorf("git merge error: %s (%v)", string(out), err)
	}
	return nil
}

// FormatStatus renders a formatted overview of the factory task queue.
func (f *Factory) FormatStatus() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var b strings.Builder
	b.WriteString("## Mímir Level 3 Autonomous Software Factory\n")
	b.WriteString(fmt.Sprintf("- **Total Tasks**: %d\n", len(f.tasks)))
	b.WriteString(fmt.Sprintf("- **Target Branch**: `%s`\n", f.cfg.TargetBranch))
	b.WriteString(fmt.Sprintf("- **Worktree Base**: `%s`\n\n", f.worktreeRoot))

	if len(f.taskOrder) == 0 {
		b.WriteString("No tasks currently queued. Submit tasks with `mimir swarm add \"Title\" \"Description\"`.\n")
		return b.String()
	}

	for _, id := range f.taskOrder {
		t := f.tasks[id]
		stateBadge := strings.ToUpper(string(t.State))
		b.WriteString(fmt.Sprintf("* **[%s] %s (`%s`)**\n", stateBadge, t.Title, t.ID))
		b.WriteString(fmt.Sprintf("  - State: `%s` | Assigned To: `%s` | Branch: `%s`\n", t.State, t.AssignedTo, t.Branch))
		if t.Error != "" {
			b.WriteString(fmt.Sprintf("  - Error: **%s**\n", t.Error))
		}
		if len(t.Artifacts) > 0 {
			b.WriteString(fmt.Sprintf("  - Deliverables: %s\n", strings.Join(t.Artifacts, ", ")))
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
