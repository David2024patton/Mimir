package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// HookType identifies a lifecycle execution point in the agent harness.
type HookType string

const (
	HookPrePrompt HookType = "pre_prompt" // Before prompt construction / LLM invocation
	HookPreTool   HookType = "pre_tool"   // Before a tool executes
	HookPostTool  HookType = "post_tool"  // Immediately after a tool finishes execution
	HookOnStop    HookType = "on_stop"    // When the agent turn finishes (final reply emitted)
)

// HookContext provides state and metadata to hooks.
type HookContext struct {
	Type        HookType
	ToolName    string
	ToolArgs    string
	ToolResult  string
	ToolErr     string
	Path        string   // Target file if mutation tool
	TodoLines   []string // Current harness-owned todo items
	Trace       []Step   // Steps taken during the turn
	StopStatus  StopGateStatus
	InjectedMsg string // Optional feedback injected back into agent history
}

// HookHandler defines a hook callback function.
type HookHandler func(ctx context.Context, hCtx *HookContext) error

// HookRule defines an external rule from .mimir/hooks.json.
type HookRule struct {
	Type    HookType `json:"type"`
	Command string   `json:"command"` // Shell command to run
	Tool    string   `json:"tool"`    // Optional tool filter (e.g. "write_file", "edit_block")
}

// HookRunner manages and executes Level 1 lifecycle hooks.
type HookRunner struct {
	mu       sync.RWMutex
	handlers map[HookType][]HookHandler
	workDir  string
}

// NewHookRunner initializes a HookRunner with built-in Level 1 verification rules.
func NewHookRunner(workDir string) *HookRunner {
	if workDir == "" {
		workDir = "."
	}
	hr := &HookRunner{
		handlers: make(map[HookType][]HookHandler),
		workDir:  workDir,
	}
	hr.registerBuiltinHooks()
	hr.loadConfigFile()
	return hr
}

// Register adds a custom in-process hook handler for a specific lifecycle phase.
func (hr *HookRunner) Register(hType HookType, handler HookHandler) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.handlers[hType] = append(hr.handlers[hType], handler)
}

func (hr *HookRunner) registerBuiltinHooks() {
	// PostTool: Run fast syntax check / formatting on modified source files.
	hr.Register(HookPostTool, func(ctx context.Context, hCtx *HookContext) error {
		if hCtx.ToolName != "write_file" && hCtx.ToolName != "edit_block" {
			return nil
		}
		target := extractTargetFile(hCtx.ToolArgs)
		if target == "" {
			return nil
		}
		fullPath := filepath.Join(hr.workDir, target)
		if !filepath.IsAbs(target) {
			fullPath = filepath.Clean(fullPath)
		} else {
			fullPath = target
		}

		ext := strings.ToLower(filepath.Ext(fullPath))
		if ext == ".go" {
			if out, err := runSyntaxCheckGo(ctx, fullPath); err != nil {
				hCtx.ToolResult += fmt.Sprintf("\n[Hook Linter Warning]: gofmt syntax validation issue:\n%s", out)
			}
		}
		return nil
	})

	// OnStop: Evaluate harness stop gate and suggest verification if tasks remain unverified.
	hr.Register(HookOnStop, func(ctx context.Context, hCtx *HookContext) error {
		gate := EvaluateStopGate(hCtx.Trace, hCtx.TodoLines)
		hCtx.StopStatus = gate
		if gate.Advisory != "" {
			hCtx.InjectedMsg = gate.Advisory
		}
		return nil
	})
}

func (hr *HookRunner) loadConfigFile() {
	configPath := filepath.Join(hr.workDir, ".mimir", "hooks.json")
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	var rules []HookRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return
	}
	for _, rule := range rules {
		r := rule
		hr.Register(r.Type, func(ctx context.Context, hCtx *HookContext) error {
			if r.Tool != "" && !strings.EqualFold(r.Tool, hCtx.ToolName) {
				return nil
			}
			cmd := exec.CommandContext(ctx, "sh", "-c", r.Command)
			if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
				cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", r.Command)
			}
			cmd.Dir = hr.workDir
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			_ = cmd.Run()
			if out.Len() > 0 {
				hCtx.ToolResult += fmt.Sprintf("\n[Hook %s Output]: %s", r.Type, strings.TrimSpace(out.String()))
			}
			return nil
		})
	}
}

// Execute runs all handlers registered for the given hook type.
func (hr *HookRunner) Execute(ctx context.Context, hCtx *HookContext) error {
	if hr == nil {
		return nil
	}
	hr.mu.RLock()
	handlers := hr.handlers[hCtx.Type]
	hr.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, hCtx); err != nil {
			return err
		}
	}
	return nil
}

func extractTargetFile(rawArgs string) string {
	var args map[string]any
	if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
		return ""
	}
	if p, ok := args["path"].(string); ok && p != "" {
		return p
	}
	if p, ok := args["target"].(string); ok && p != "" {
		return p
	}
	return ""
}

func runSyntaxCheckGo(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "gofmt", "-e", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), err
	}
	return "", nil
}
