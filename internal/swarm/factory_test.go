package swarm

import (
	"strings"
	"testing"
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
