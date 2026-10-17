package tools

import (
	"context"
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
