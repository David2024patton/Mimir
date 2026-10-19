package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTodoStoreLifecycle(t *testing.T) {
	s := NewTodoStore()
	todo := s.Add("write tests")
	if todo.ID != 1 || todo.Status != TodoPending {
		t.Fatalf("unexpected todo: %+v", todo)
	}
	if !s.SetStatus(1, TodoCompleted) {
		t.Fatal("SetStatus failed")
	}
	list := s.List()
	if len(list) != 1 || list[0].Status != TodoCompleted {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func TestTodoToolRun(t *testing.T) {
	tool := &TodoTool{Store: NewTodoStore()}
	ctx := context.Background()
	if _, err := tool.Run(ctx, map[string]any{"action": "add", "content": "ship it"}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	out, err := tool.Run(ctx, map[string]any{"action": "list"})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty todo list")
	}
}

func TestDefaultRegistryHasCoreTools(t *testing.T) {
	r := Default(t.TempDir())
	for _, name := range []string{"bash", "read_file", "write_file", "list_directory", "todo", "question", "load_skill"} {
		if _, ok := r.Get(name); !ok {
			t.Errorf("default registry missing tool %q", name)
		}
	}
}

func TestRegistryForModeReadOnlyHidesWriteTools(t *testing.T) {
	r := Default(t.TempDir())
	roTools := r.ForMode("ro")
	for _, tl := range roTools {
		if IsWriteTool(tl.Name()) {
			t.Errorf("read-only mode exposed write tool %q", tl.Name())
		}
	}
}

func TestClosestBlockHint(t *testing.T) {
	content := "package main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n"
	hint := closestBlockHint(content, "func main() {\n\tfmt.Println(\"wrong\")\n}")
	if hint == "" {
		t.Fatal("expected non-empty closestBlockHint")
	}
}

func TestEditBlockPreservesFormattingOnMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	original := "line 1\n    line 2 indented\n\nline 3\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := &EditBlockTool{}
	_, err := tool.Run(context.Background(), map[string]any{
		"path":       path,
		"old_string": "line 2 with wrong text",
		"new_string": "replaced line 2",
	})
	if err == nil {
		t.Fatal("expected error when old_string does not match")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != original {
		t.Fatalf("file content was modified on mismatch error!\nExpected:\n%q\nGot:\n%q", original, string(after))
	}
}

func TestWorkspaceConfinementPreventEscapes(t *testing.T) {
	wsDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("sensitive info"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := Default(wsDir)

	// 1. read_file outside workspace should fail
	rf, ok := r.Get("read_file")
	if !ok {
		t.Fatal("read_file not found")
	}
	if _, err := rf.Run(context.Background(), map[string]any{"path": outsideFile}); err == nil {
		t.Fatal("expected security error reading file outside workspace")
	}

	// 2. read_file with relative path escaping workspace should fail
	if _, err := rf.Run(context.Background(), map[string]any{"path": "../secret.txt"}); err == nil {
		t.Fatal("expected security error for path traversal ../")
	}

	// 3. write_file outside workspace should fail
	wf, ok := r.Get("write_file")
	if !ok {
		t.Fatal("write_file not found")
	}
	if _, err := wf.Run(context.Background(), map[string]any{"path": outsideFile, "content": "malicious"}); err == nil {
		t.Fatal("expected security error writing file outside workspace")
	}

	// 4. write_file inside workspace should succeed
	insideFile := filepath.Join(wsDir, "ok.txt")
	if _, err := wf.Run(context.Background(), map[string]any{"path": "ok.txt", "content": "safe"}); err != nil {
		t.Fatalf("failed to write inside workspace: %v", err)
	}
	data, err := os.ReadFile(insideFile)
	if err != nil || string(data) != "safe" {
		t.Fatalf("unexpected content: %s, %v", string(data), err)
	}
}

// N9: a symlink inside the workspace pointing outside must not pass the
// confinement check for read, write, or list; a symlink that stays inside
// the workspace keeps working.
func TestWorkspaceConfinementRejectsSymlinkEscape(t *testing.T) {
	wsDir := t.TempDir()
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("sensitive info"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(wsDir, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// Positive control: an in-workspace symlink target.
	realDir := filepath.Join(wsDir, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "ok.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(wsDir, "alias")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	r := Default(wsDir)
	ctx := context.Background()

	rf, _ := r.Get("read_file")
	if _, err := rf.Run(ctx, map[string]any{"path": "escape/secret.txt"}); err == nil {
		t.Fatal("expected security error reading through symlink escape")
	}
	if out, err := rf.Run(ctx, map[string]any{"path": "alias/ok.txt"}); err != nil || out != "inside" {
		t.Fatalf("expected in-workspace symlink read to succeed, got %q, %v", out, err)
	}

	wf, _ := r.Get("write_file")
	if _, err := wf.Run(ctx, map[string]any{"path": "escape/new.txt", "content": "malicious"}); err == nil {
		t.Fatal("expected security error writing through symlink escape")
	}
	if _, err := os.Stat(filepath.Join(outsideDir, "new.txt")); !os.IsNotExist(err) {
		t.Fatal("write through symlink escape created a file outside the workspace")
	}
	// Writing a not-yet-existing file through an in-workspace path is fine.
	if _, err := wf.Run(ctx, map[string]any{"path": "alias/new.txt", "content": "fine"}); err != nil {
		t.Fatalf("expected write via in-workspace symlink parent to succeed: %v", err)
	}

	lf, _ := r.Get("list_directory")
	if _, err := lf.Run(ctx, map[string]any{"path": "escape"}); err == nil {
		t.Fatal("expected security error listing through symlink escape")
	}
}
