package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const maxShellOutputBytes = 32 * 1024 // 32 KB context cap (sausheong/harness pattern)

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
	result := capShellOutput(out.String())
	if err != nil {
		return result, fmt.Errorf("command failed: %w", err)
	}
	return result, nil
}

func capShellOutput(s string) string {
	if len(s) <= maxShellOutputBytes {
		return s
	}
	spillFile, err := os.CreateTemp("", "mimir-bash-*.log")
	spillNote := ""
	if err == nil {
		_, _ = spillFile.WriteString(s)
		_ = spillFile.Close()
		spillNote = fmt.Sprintf(" Full output (%d bytes) saved to %s.", len(s), filepath.Clean(spillFile.Name()))
	}
	head := s[:16*1024]
	tail := s[len(s)-12*1024:]
	return fmt.Sprintf("%s\n\n... [output truncated: %d bytes omitted.%s] ...\n\n%s",
		head, len(s)-28*1024, spillNote, tail)
}
