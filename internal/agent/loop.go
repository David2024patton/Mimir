package agent

import (
	"context"
	"fmt"

	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// State is the agent-loop state.
type State string

const (
	StateIdle             State = "idle"
	StateRunning          State = "running"
	StateAwaitingApproval State = "awaiting_approval"
	StateCompacting       State = "compacting"
	StateDone             State = "done"
	StateError            State = "error"
)

// Loop is the agent loop (E2.3): assemble -> infer (stream) -> act (tools) -> verify.
type Loop struct {
	Providers *llm.Registry
	Tools     *tools.Registry
}

func NewLoop(providers *llm.Registry, tools *tools.Registry) *Loop {
	return &Loop{Providers: providers, Tools: tools}
}

// Run executes one turn, emitting stream events to emit.
func (l *Loop) Run(ctx context.Context, model string, messages []llm.Message, emit func(llm.StreamEvent)) error {
	provider, err := l.Providers.Route(model)
	if err != nil {
		return err
	}
	req := llm.GenerateRequest{Model: model, Messages: messages, Stream: true}
	for _, t := range l.Tools.All() {
		req.Tools = append(req.Tools, llm.ToolSchema{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		})
	}
	events, err := provider.Stream(ctx, req)
	if err != nil {
		return err
	}
	for ev := range events {
		emit(ev)
		// TODO(E2.3): on a tool_call event, run the policy gate, execute the tool,
		// feed the result back as a tool message, and loop until the model stops.
		if ev.Type == "error" {
			return fmt.Errorf("stream error: %w", ev.Err)
		}
	}
	return nil
}
