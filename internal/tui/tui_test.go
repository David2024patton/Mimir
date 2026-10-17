package tui

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

type fakeTUIProvider struct{}

func (f *fakeTUIProvider) ID() string { return "ollama" }
func (f *fakeTUIProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{Content: "Skål from Mímir!", FinishReason: "stop"}, nil
}
func (f *fakeTUIProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 2)
	ch <- llm.StreamEvent{Delta: "Skål from Mímir!"}
	ch <- llm.StreamEvent{Done: true, Usage: llm.Usage{PromptTokens: 12, CompletionTokens: 5}}
	close(ch)
	return ch, nil
}

func TestTUISlashCommandsAndChatTurn(t *testing.T) {
	store, err := cortex.NewMemoryStoreAt(filepath.Join(t.TempDir(), "cortex.json"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	reg := tools.Default(t.TempDir())
	tracker := llm.NewTracker()
	ag := agent.New(agent.Config{
		Provider: &fakeTUIProvider{},
		Tools:    reg,
		Cortex:   store,
		Model:    "qwen3:8b",
		Usage:    tracker,
		Local:    true,
	})

	input := strings.Join([]string{
		"/mode ro",
		"/effort off",
		"/decide refactor the polyglot architecture",
		"/persona odin",
		"hello mimir",
		"/quit",
	}, "\n") + "\n"

	var out bytes.Buffer
	err = Run(context.Background(), Options{
		Version:   "0.2.0-test",
		Agent:     ag,
		Tools:     reg,
		Cortex:    store,
		Usage:     tracker,
		StoreDesc: "file-backed test",
		In:        strings.NewReader(input),
		Out:       &out,
	})
	if err != nil {
		t.Fatalf("TUI Run failed: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Mímir 0.2.0-test",
		"Mode -> ro",
		"Reasoning effort -> off",
		"Huginn-Decider (liquid/d1-omni-600m)",
		"Persona=Odin",
		"Persona active: odin",
		"Skål from Mímir!",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("TUI output missing %q.\nFull output:\n%s", want, got)
		}
	}
}

func TestPromptQuestionsSelector(t *testing.T) {
	qs := []tools.Question{
		{
			Header:   "Database",
			Question: "Which Cortex backend do you want?",
			Options: []tools.Option{
				{Label: "File JSON", Description: "Zero-setup local file"},
				{Label: "SurrealDB", Description: "Hybrid vector + FTS"},
			},
		},
	}
	var out bytes.Buffer
	ans, err := PromptQuestions(context.Background(), strings.NewReader("2\n"), &out, qs)
	if err != nil {
		t.Fatalf("PromptQuestions failed: %v", err)
	}
	if !strings.Contains(ans, "Database: SurrealDB") {
		t.Errorf("got answer %q, want 'Database: SurrealDB'", ans)
	}
}
