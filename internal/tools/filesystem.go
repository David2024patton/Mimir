package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool reads a file's contents (E3.2).
type ReadFileTool struct {
	WorkDir string
}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Read the contents of a file within the workspace."
}
func (t *ReadFileTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string"},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Run(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	resolved, err := resolveWithinWorkDir(t.WorkDir, path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFileTool writes content to a file (E3.2).
type WriteFileTool struct {
	WorkDir string
}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write content to a file within the workspace (creates or overwrites)."
}
func (t *WriteFileTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string"},
			"content": map[string]any{"type": "string"},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Run(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	resolved, err := resolveWithinWorkDir(t.WorkDir, path)
	if err != nil {
		return "", err
	}
	if dir := filepath.Dir(resolved); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path), nil
}

// ListDirectoryTool lists a directory's entries (E3.2).
type ListDirectoryTool struct {
	WorkDir string
}

func (t *ListDirectoryTool) Name() string { return "list_directory" }
func (t *ListDirectoryTool) Description() string {
	return "List files and directories at a path within the workspace."
}
func (t *ListDirectoryTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string"},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirectoryTool) Run(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	resolved, err := resolveWithinWorkDir(t.WorkDir, path)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, e := range entries {
		suffix := ""
		if e.IsDir() {
			suffix = "/"
		}
		fmt.Fprintf(&b, "%s%s\n", e.Name(), suffix)
	}
	return b.String(), nil
}

func resolveWithinWorkDir(workDir, targetPath string) (string, error) {
	if strings.TrimSpace(targetPath) == "" {
		targetPath = "."
	}
	if workDir == "" {
		workDir = "."
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	// Resolve symlinks in the root itself (e.g. /tmp aliases) so the
	// containment checks below compare real paths on both sides.
	if resolvedRoot, err := filepath.EvalSymlinks(absWorkDir); err == nil {
		absWorkDir = resolvedRoot
	}
	var candidate string
	if filepath.IsAbs(targetPath) {
		candidate = filepath.Clean(targetPath)
	} else {
		candidate = filepath.Clean(filepath.Join(absWorkDir, targetPath))
	}
	if !pathWithinDir(absWorkDir, candidate) {
		return "", fmt.Errorf("security violation: path %q escapes workspace boundary %q", targetPath, workDir)
	}
	// Symlink check (N9): a purely lexical check lets a symlink inside the
	// workspace point anywhere on disk. Resolve the longest existing
	// prefix of the candidate and re-check containment on the real path.
	// A not-yet-existing final component is fine: its parent is resolved.
	resolved, err := resolveExistingPrefix(candidate)
	if err != nil {
		return "", fmt.Errorf("security violation: cannot resolve path %q: %w", targetPath, err)
	}
	if !pathWithinDir(absWorkDir, resolved) {
		return "", fmt.Errorf("security violation: path %q escapes workspace boundary %q via symlink", targetPath, workDir)
	}
	return resolved, nil
}

// pathWithinDir reports whether candidate is dir itself or lives under it.
func pathWithinDir(dir, candidate string) bool {
	rel, err := filepath.Rel(dir, candidate)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolveExistingPrefix resolves symlinks on the longest prefix of path
// that exists on disk and reattaches the remaining (not yet existing)
// components unchanged.
func resolveExistingPrefix(path string) (string, error) {
	p := path
	var suffix []string
	for {
		if _, err := os.Lstat(p); err == nil {
			resolved, err := filepath.EvalSymlinks(p)
			if err != nil {
				return "", err
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return resolved, nil
		}
		parent := filepath.Dir(p)
		if parent == p {
			return path, nil
		}
		suffix = append(suffix, filepath.Base(p))
		p = parent
	}
}
