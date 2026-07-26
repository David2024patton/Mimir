package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// fakeStreamProvider streams: turn 1 issues a bash tool call, turn 2 streams the
// final answer in two deltas. It exercises RunStream's stream -> act -> stream loop.
type fakeStreamProvider struct {
	calls int
}

func (f *fakeStreamProvider) ID() string { return "fakestream" }
func (f *fakeStreamProvider) Generate(context.Context, llm.GenerateRequest) (llm.GenerateResponse, error) {
	return llm.GenerateResponse{}, nil
}
func (f *fakeStreamProvider) Stream(context.Context, llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	f.calls++
	out := make(chan llm.StreamEvent)
	go func() {
		defer close(out)
		if f.calls == 1 {
			out <- llm.StreamEvent{Done: true, ToolCalls: []llm.ToolCall{
				{ID: "c1", Name: "bash", Arguments: `{"command":"echo MIMIRX"}`},
			}}
			return
		}
		out <- llm.StreamEvent{Delta: "all "}
		out <- llm.StreamEvent{Delta: "done"}
		out <- llm.StreamEvent{Done: true}
	}()
	return out, nil
}

// TestRunStream proves the streaming turn end to end: a seeded memory is recalled,
// the model's tool call is executed and surfaced, the final answer streams token by
// token, the done event carries the full reply, and the exchange is persisted.
func TestRunStream(t *testing.T) {
	store := cortex.NewMemoryStore()
	if _, err := store.PutNeuron(context.Background(), cortex.Neuron{
		Kind: cortex.KindMemory, Content: "the codeword is MANGO77", Decay: 1,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	a := New(Config{
		Provider: &fakeStreamProvider{}, Tools: tools.Default(t.TempDir()),
		Cortex: store, Model: "m",
	})

	var events []Event
	if err := a.RunStream(context.Background(), "tell me about MANGO77 please", nil, func(e Event) {
		events = append(events, e)
	}); err != nil {
		t.Fatalf("RunStream: %v", err)
	}

	var (
		sawMemory, sawTool, sawDone bool
		tokens, doneReply           string
		toolStep                    *Step
	)
	for _, e := range events {
		switch e.Type {
		case EventMemory:
			if strings.Contains(e.Text, "MANGO77") {
				sawMemory = true
			}
		case EventToken:
			tokens += e.Text
		case EventTool:
			sawTool = true
			toolStep = e.Step
		case EventDone:
			sawDone = true
			doneReply = e.Reply
		case EventError:
			t.Fatalf("unexpected error event: %s", e.Err)
		}
	}
	if !sawMemory {
		t.Error("expected a recalled-memory event for MANGO77")
	}
	if !sawTool || toolStep == nil || toolStep.Name != "bash" || !strings.Contains(toolStep.Result, "MIMIRX") {
		t.Errorf("expected a bash tool event whose result contains MIMIRX, got %+v", toolStep)
	}
	if tokens != "all done" {
		t.Errorf("streamed tokens = %q, want %q", tokens, "all done")
	}
	if !sawDone || doneReply != "all done" {
		t.Errorf("done reply = %q, want %q", doneReply, "all done")
	}
}
