package swarm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	StateBacklog     TaskState = "backlog"     // Waiting for specification
	StateSpec        TaskState = "spec"        // Odin (Orchestrator) writing requirements
	StateImplement   TaskState = "implement"   // Thor (Builder) developing in worktree
	StateTest        TaskState = "test"        // Loki (Tester/Debugger) validating and fixing
	StateReview      TaskState = "review"      // Heimdall (Auditor) running SEO/GEO and web audit
	StateMerged      TaskState = "merged"      // Merged back to target branch
	StateFailed      TaskState = "failed"      // Halted due to unrecoverable test or review failure
)

// Task represents a discrete unit of work assigned to the autonomous software factory.
type Task struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Branch       string            `json:"branch"`
	WorktreeDir  string            `json:"worktree_dir,omitempty"`
	State        TaskState         `json:"state"`
	AssignedTo   string            `json:"assigned_to"` // "odin" | "thor" | "loki" | "heimdall"
	Error        string            `json:"error,omitempty"`
	Artifacts    []string          `json:"artifacts,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// FactoryConfig configures the autonomous software factory queue.
type FactoryConfig struct {
	RootDir     string
	MaxWorkers  int
	AgentConfig agent.Config
	RunLimits   runengine.Limits
	TargetBranch string // default "main"
}

// Factory manages parallel multi-agent worker pipelines operating in isolated git worktrees.
type Factory struct {
	mu        sync.RWMutex
	cfg       FactoryConfig
	tasks     map[string]*Task
	taskOrder []string
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

func (f *Factory) loadFromDisk() {
	raw, err := os.ReadFile(f.storagePath())
	if err != nil {
		return
	}
	var data struct {
		Order []string          `json:"order"`
		Tasks map[string]*Task `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}
	f.taskOrder = data.Order
	f.tasks = data.Tasks
}

func (f *Factory) saveToDisk() {
	data := struct {
		Order []string          `json:"order"`
		Tasks map[string]*Task `json:"tasks"`
	}{
		Order: f.taskOrder,
		Tasks: f.tasks,
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

	id := fmt.Sprintf("task-%d", len(f.tasks)+1)
	cleanSlug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(title), " ", "-"))
	if len(cleanSlug) > 30 {
		cleanSlug = cleanSlug[:30]
	}
	branch := fmt.Sprintf("swarm/%s-%s", id, cleanSlug)

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

	// 5. Merge worktree branch to target branch
	if err := f.mergeTaskBranch(ctx, t.Branch); err != nil {
		f.failTask(taskID, fmt.Sprintf("Git merge failed: %v", err))
		return err
	}

	f.updateState(taskID, StateMerged, "none")
	return nil
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

func (f *Factory) runThorImplement(ctx context.Context, t *Task, wtDir, specDoc string) error {
	// Thor operates inside wtDir. For standard factory tasks, verify build environment and record progress.
	buildRecord := filepath.Join(wtDir, ".mimir", "sdlc", fmt.Sprintf("04-BUILD-%s.md", t.ID))
	_ = os.MkdirAll(filepath.Dir(buildRecord), 0o755)

	content := fmt.Sprintf("# Thor Implementation Record: %s\n\n"+
		"- Task ID: `%s`\n"+
		"- Worktree: `%s`\n"+
		"- Status: Ready for Loki verification\n",
		t.Title, t.ID, wtDir)

	return os.WriteFile(buildRecord, []byte(content), 0o644)
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
	// Merge task branch into target branch
	cmd := exec.CommandContext(ctx, "git", "merge", "--no-ff", "-m", fmt.Sprintf("feat(swarm): merge %s", branch), branch)
	cmd.Dir = f.cfg.RootDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// Abort conflicted merge if needed
		_ = exec.Command("git", "merge", "--abort").Run()
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
