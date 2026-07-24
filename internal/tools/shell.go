package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// ShellTool runs a shell command and returns its output (E3.1).
type ShellTool struct {
	WorkDir string
}

func (t *ShellTool) Name() string        { return "bash" }
func (t *ShellTool) Description() string { return "Run a shell command and return stdout/stderr." }
func (t *ShellTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{"type": "string", "description": "The command to run"},
		},
		"required": []string{"command"},
	}
}

func (t *ShellTool) Run(ctx context.Context, args map[string]any) (string, error) {
	command, _ := args["command"].(string)
	if command == "" {
		return "", fmt.Errorf("command is required")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", command)
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}
	if t.WorkDir != "" {
		cmd.Dir = t.WorkDir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	result := out.String()
	if err != nil {
		return result, fmt.Errorf("command failed: %w", err)
	}
	return result, nil
}
