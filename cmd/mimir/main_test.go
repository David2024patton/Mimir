package main

import (
	"os"
	"testing"
)

// TestReachesPromptRun pins the dispatch prediction the approver wiring
// relies on (FIX 2): only a non-empty one-shot prompt (plain args, a
// chat/trace mode word plus text, or -p) reaches the prompt path; leading
// subcommands return before it, and a bare mode word is not a prompt.
func TestReachesPromptRun(t *testing.T) {
	cases := []struct {
		name       string
		positional []string
		promptFlag string
		want       bool
	}{
		{"plain prompt args", []string{"fix", "the", "bug"}, "", true},
		{"chat mode with prompt", []string{"chat", "hello"}, "", true},
		{"trace mode with prompt", []string{"trace", "hello"}, "", true},
		{"chat mode alone", []string{"chat"}, "", false},
		{"trace mode alone", []string{"trace"}, "", false},
		{"no args at all", nil, "", false},
		{"prompt flag only", nil, "hello", true},
		{"serve subcommand", []string{"serve"}, "", false},
		{"serve with prompt flag", []string{"serve"}, "hello", false},
		{"swarm subcommand", []string{"swarm", "run", "task-1"}, "", false},
		{"pair subcommand", []string{"pair"}, "", false},
		{"review subcommand", []string{"review"}, "", false},
	}
	for _, tc := range cases {
		if got := reachesPromptRun(tc.positional, tc.promptFlag); got != tc.want {
			t.Errorf("%s: reachesPromptRun(%v, %q) = %v, want %v",
				tc.name, tc.positional, tc.promptFlag, got, tc.want)
		}
	}
}

// TestApproverForRun pins the one-stdin-owner rule (FIX 2): the stdin
// terminal approver exists only for a one-shot prompt run at a terminal.
// TUI, serve, and --web runs get nil, so RequireApproval denies there
// fail-closed instead of prompting a terminal nobody is watching (or a
// second buffered reader fighting the TUI's own stdin scanner).
func TestApproverForRun(t *testing.T) {
	if approverForRun(true, true) == nil {
		t.Error("one-shot run at a terminal: approver = nil, want the terminal approver")
	}
	if approverForRun(true, false) != nil {
		t.Error("one-shot run without a terminal: approver non-nil, want nil")
	}
	if approverForRun(false, true) != nil {
		t.Error("TUI/serve/web run at a terminal: approver non-nil, want nil (TUI owns stdin; HTTP callers have no terminal)")
	}
	if approverForRun(false, false) != nil {
		t.Error("non-one-shot run without a terminal: approver non-nil, want nil")
	}
}

// swapStdin replaces os.Stdin with a pipe preloaded with input (write end
// already closed, so reads hit EOF after the content) and resets the shared
// approval reader so the approver reads from the new stdin.
func swapStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("pipe write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("pipe close: %v", err)
	}
	old := os.Stdin
	os.Stdin = r
	approvalMu.Lock()
	approvalReader = nil
	approvalMu.Unlock()
	t.Cleanup(func() {
		os.Stdin = old
		approvalMu.Lock()
		approvalReader = nil
		approvalMu.Unlock()
		_ = r.Close()
	})
}

// TestTerminalApproverEOFDenies proves the approver fails closed on EOF:
// no readable answer is a denial, never an approval.
func TestTerminalApproverEOFDenies(t *testing.T) {
	swapStdin(t, "")
	if terminalApprover("bash", "{}") {
		t.Error("EOF on stdin: approver returned true, want deny")
	}
}

// TestTerminalApproverAnswers proves only an explicit y/yes line approves.
func TestTerminalApproverAnswers(t *testing.T) {
	swapStdin(t, "y\n")
	if !terminalApprover("bash", "{}") {
		t.Error("explicit y: approver returned false, want approve")
	}
	swapStdin(t, "n\n")
	if terminalApprover("bash", "{}") {
		t.Error("explicit n: approver returned true, want deny")
	}
	swapStdin(t, "whatever\n")
	if terminalApprover("bash", "{}") {
		t.Error("non-answer line: approver returned true, want deny")
	}
}
