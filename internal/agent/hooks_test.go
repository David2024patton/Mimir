package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHookRunner_RegistrationAndExecution(t *testing.T) {
	hr := NewHookRunner(t.TempDir())
	called := false
	hr.Register(HookPreTool, func(ctx context.Context, hCtx *HookContext) error {
		if hCtx.ToolName == "test_tool" {
			called = true
		}
		return nil
	})

	hCtx := &HookContext{
		Type:     HookPreTool,
		ToolName: "test_tool",
	}
	if err := hr.Execute(context.Background(), hCtx); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !called {
		t.Fatalf("expected hook to be called")
	}
}

func TestHookRunner_PostToolCatchesGoSyntaxError(t *testing.T) {
	dir := t.TempDir()
	badGoFile := filepath.Join(dir, "broken.go")
	if err := os.WriteFile(badGoFile, []byte("package broken\nfunc foo( { nope"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hr := NewHookRunner(dir)
	hCtx := &HookContext{
		Type:       HookPostTool,
		ToolName:   "write_file",
		ToolArgs:   `{"path": "broken.go"}`,
		ToolResult: "file written successfully",
	}

	if err := hr.Execute(context.Background(), hCtx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(hCtx.ToolResult, "[Hook Linter Warning]") {
		t.Fatalf("expected linter warning in ToolResult, got:\n%s", hCtx.ToolResult)
	}
}

func TestHookRunner_OnStopEvaluatesGate(t *testing.T) {
	hr := NewHookRunner(t.TempDir())
	hCtx := &HookContext{
		Type: HookOnStop,
		TodoLines: []string{
			"- #1 [in_progress] finish hooks",
		},
		Trace: []Step{
			{Name: "write_file", Args: `{"path":"main.go"}`},
		},
	}

	if err := hr.Execute(context.Background(), hCtx); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if hCtx.InjectedMsg == "" || !strings.Contains(hCtx.InjectedMsg, "Harness OnStop Gate Advisory") {
		t.Fatalf("expected stop gate advisory, got %q", hCtx.InjectedMsg)
	}
}
