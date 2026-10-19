package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/decision"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/policy"
	"github.com/David2024patton/Mimir/internal/sdlc"
	"github.com/David2024patton/Mimir/internal/tools"
)

// maxToolSteps caps the infer->act loop so a misbehaving model can't spin forever
// (doom-loop protection, F4.3).
const maxToolSteps = 20

// Permission modes (inspired by smolcoder ro / edit / bypass).
const (
	ModeReadOnly = "ro"
	ModeEdit     = "edit"
	ModeBypass   = "bypass"
)

// Config wires an agent's dependencies.
type Config struct {
	Provider      llm.Provider
	Tools         *tools.Registry
	Cortex        cortex.Store
	Model         string
	Usage         *llm.Tracker      // optional: records token usage for the dashboard (E69)
	Decider       *decision.Decider // System-One Decision Engine (defaults to Liquid AI d1-omni-600M)
	SDLC          *sdlc.Manager     // 7-Phase SDLC State Machine & Phase-Scoped Prompt Loader
	Local         bool              // whether the provider is local (local tokens cost $0)
	Mode          string            // "ro" | "edit" | "bypass" (default "edit")
	Effort        string            // "" | "off" | "low" | "medium" | "high"
	SmolPolicy    string            // "auto" (default) | "on" | "off"
	ApprovalFunc  func(toolName, rawArgs string) bool
	Gate          *policy.Gate // Fail-closed execution policy gate (F2/ADR-009)
	Hooks         *HookRunner  // Level 1 Lifecycle Hooks
	ContextBudget int          // Max context token budget (defaults to 8192 in SmolMode, 32768 standard)
}

// Step records one tool execution in a turn (used by RunTrace / the trace CLI).
type Step struct {
	Name   string `json:"name"`
	Args   string `json:"args"`
	Result string `json:"result"`
	Err    string `json:"err,omitempty"`
}

// Result is the full outcome of a turn: the final reply, the tool trace, and the
// memories that were recalled into the prompt (so memory is observable/testable).
type Result struct {
	Reply    string
	Trace    []Step
	Recalled []string
}

// Event kinds emitted by RunStream (F2.1 streaming).
const (
	EventMemory = "memory" // recalled memory (Text = content)
	EventToken  = "token"  // a delta of the model's reply (Text = delta)
	EventTool   = "tool"   // a tool execution (Step)
	EventSteer  = "steer"  // mid-turn user steering injected
	EventDone   = "done"   // turn complete (Reply = final answer)
	EventError  = "error"  // turn failed (Err)
)

// Event is one streaming update from a turn, surfaced to live UIs by RunStream.
type Event struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Step  *Step  `json:"step,omitempty"`
	Reply string `json:"reply,omitempty"`
	Err   string `json:"err,omitempty"`
}

// Agent runs the gather -> infer -> act -> verify loop.
type Agent struct {
	mu      sync.RWMutex
	cfg     Config
	steerCh chan string
}

// New returns an Agent wired to the given dependencies.
func New(cfg Config) *Agent {
	if cfg.Mode == "" {
		cfg.Mode = ModeEdit
	}
	if cfg.SmolPolicy == "" {
		cfg.SmolPolicy = SmolAuto
	}
	if cfg.Decider == nil {
		cfg.Decider = decision.New("", "", decision.PrimaryDecisionModel)
	}
	if cfg.SDLC == nil {
		if cfg.Tools != nil {
			if st, ok := cfg.Tools.Get("sdlc"); ok {
				if sdlcTool, ok := st.(*tools.SDLCTool); ok && sdlcTool.Manager != nil {
					cfg.SDLC = sdlcTool.Manager
				}
			}
		}
		if cfg.SDLC == nil {
			cfg.SDLC = sdlc.NewManager(".")
		}
	}
	if cfg.Hooks == nil {
		cfg.Hooks = NewHookRunner(".")
	}
	return &Agent{
		cfg:     cfg,
		steerCh: make(chan string, 16),
	}
}

// Hooks returns Mímir's Level 1 lifecycle hook runner.
func (a *Agent) Hooks() *HookRunner {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Hooks
}

// Decider returns Mímir's built-in System-One Decision Engine (Liquid AI d1-omni-600M).
func (a *Agent) Decider() *decision.Decider {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Decider
}

// SDLC returns Mímir's 7-Phase SDLC State Machine & Phase-Scoped Prompt Loader.
func (a *Agent) SDLC() *sdlc.Manager {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.SDLC
}

// SmolMode reports whether smolcoder tiny-model optimizations are currently active
// (either forced on or auto-activated by DetectSmolModel).
func (a *Agent) SmolMode() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	switch a.cfg.SmolPolicy {
	case SmolOn:
		return true
	case SmolOff:
		return false
	default:
		pid := ""
		if a.cfg.Provider != nil {
			pid = a.cfg.Provider.ID()
		}
		return DetectSmolModel(pid, a.cfg.Model, a.cfg.Local)
	}
}

// SmolPolicy returns the current SmolMode policy ("auto", "on", "off").
func (a *Agent) SmolPolicy() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.SmolPolicy == "" {
		return SmolAuto
	}
	return a.cfg.SmolPolicy
}

// SetSmolPolicy sets the SmolMode policy ("auto", "on", "off").
func (a *Agent) SetSmolPolicy(policy string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case SmolOn:
		a.cfg.SmolPolicy = SmolOn
	case SmolOff:
		a.cfg.SmolPolicy = SmolOff
	default:
		a.cfg.SmolPolicy = SmolAuto
	}
}

// Mode returns the current permission mode ("ro", "edit", "bypass").
func (a *Agent) Mode() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.Mode == "" {
		return ModeEdit
	}
	return a.cfg.Mode
}

// SetMode sets the active permission mode ("ro", "edit", "bypass").
func (a *Agent) SetMode(mode string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch mode {
	case ModeReadOnly, ModeEdit, ModeBypass:
		a.cfg.Mode = mode
	}
}

// CycleMode cycles permission modes: edit -> bypass -> ro -> edit (Shift+Tab in TUI).
func (a *Agent) CycleMode() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch a.cfg.Mode {
	case ModeReadOnly:
		a.cfg.Mode = ModeEdit
	case ModeEdit:
		a.cfg.Mode = ModeBypass
	default:
		a.cfg.Mode = ModeReadOnly
	}
	return a.cfg.Mode
}

// Effort returns the current reasoning effort ("", "off", "low", "medium", "high").
func (a *Agent) Effort() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Effort
}

// SetEffort updates the reasoning effort level.
func (a *Agent) SetEffort(effort string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Effort = strings.ToLower(strings.TrimSpace(effort))
}

// SetProvider hot-swaps the active LLM provider and model at runtime.
func (a *Agent) SetProvider(p llm.Provider, model string, local bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Provider = p
	a.cfg.Model = model
	a.cfg.Local = local
}

// ModelInfo returns the current provider ID, model name, and local flag.
func (a *Agent) ModelInfo() (string, string, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	pid := "none"
	if a.cfg.Provider != nil {
		pid = a.cfg.Provider.ID()
	}
	return pid, a.cfg.Model, a.cfg.Local
}

// Steer queues a mid-turn user steering instruction to be injected cleanly at the
// next joined tool boundary without aborting the turn (sausheong/harness pattern).
func (a *Agent) Steer(msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	select {
	case a.steerCh <- msg:
	default:
	}
}

func (a *Agent) drainSteer() []string {
	var out []string
	for {
		select {
		case m := <-a.steerCh:
			out = append(out, m)
		default:
			return out
		}
	}
}

// phaseToolFilter returns the active SDLC phase's allowed tool set when a
// methodology with phase-scoped tools is active (waterfall), or nil when no
// phase restriction applies. It is the single source of the computation used
// both to filter advertised schemas and to enforce the set at execution time.
func (a *Agent) phaseToolFilter() map[string]bool {
	a.mu.RLock()
	sdlcMgr := a.cfg.SDLC
	a.mu.RUnlock()
	if sdlcMgr != nil && sdlcMgr.Methodology() == sdlc.MethodWaterfall {
		return sdlcMgr.AllowedToolSet()
	}
	return nil
}

// schemas converts the registered tools visible in the current mode into model-facing tool schemas.
// When SDLC Waterfall methodology is active, it also enforces the active phase's AllowedToolSet.
func (a *Agent) schemas() []llm.ToolSchema {
	a.mu.RLock()
	reg := a.cfg.Tools
	mode := a.cfg.Mode
	a.mu.RUnlock()
	if reg == nil {
		return nil
	}
	phaseFilter := a.phaseToolFilter()
	var out []llm.ToolSchema
	for _, t := range reg.ForMode(mode) {
		if phaseFilter != nil && !phaseFilter[t.Name()] {
			continue
		}
		out = append(out, llm.ToolSchema{
			Name: t.Name(), Description: t.Description(), Parameters: t.Parameters(),
		})
	}
	return out
}

func (a *Agent) todoLines() []string {
	a.mu.RLock()
	reg := a.cfg.Tools
	a.mu.RUnlock()
	if reg == nil {
		return nil
	}
	t, ok := reg.Get("todo")
	if !ok {
		return nil
	}
	tt, ok := t.(*tools.TodoTool)
	if !ok || tt.Store == nil {
		return nil
	}
	var lines []string
	for _, item := range tt.Store.List() {
		lines = append(lines, fmt.Sprintf("- #%d [%s] %s", item.ID, item.Status, item.Content))
	}
	return lines
}

func (a *Agent) executeToolCall(ctx context.Context, tc llm.ToolCall) Step {
	repairedArgs := llm.RepairJSON(tc.Arguments)
	st := Step{Name: tc.Name, Args: repairedArgs}

	a.mu.RLock()
	reg := a.cfg.Tools
	mode := a.cfg.Mode
	approve := a.cfg.ApprovalFunc
	gate := a.cfg.Gate
	a.mu.RUnlock()

	var args map[string]any
	if repairedArgs != "" {
		_ = json.Unmarshal([]byte(repairedArgs), &args)
	}

	// 0. SDLC phase enforcement at execution time (F5): schemas() only filters
	// which tools are advertised, so a model that emits a call anyway (or a
	// direct caller) must still be rejected when the active phase's allowed
	// tool set does not include the tool.
	if phaseFilter := a.phaseToolFilter(); phaseFilter != nil && !phaseFilter[tc.Name] {
		st.Err = fmt.Sprintf("blocked by SDLC phase policy: tool %q is not in the active phase's allowed tool set", tc.Name)
		st.Result = st.Err
		return st
	}

	// 1. Fail-closed policy gate check (F2/F3/ADR-009)
	approved := false
	if gate != nil {
		dec := gate.Evaluate(policy.Request{Tool: tc.Name, Args: args})
		switch dec {
		case policy.Deny:
			st.Err = fmt.Sprintf("execution of tool %q blocked by policy gate (deny)", tc.Name)
			st.Result = st.Err
			return st
		case policy.RequireApproval:
			// Bypass mode is the user's explicit opt-out of approval prompts:
			// RequireApproval verdicts are waived in bypass mode. Deny verdicts
			// are NOT waived (handled above) and still block execution.
			if mode == ModeBypass {
				// Waived by explicit user opt-out; proceed without approval.
			} else if approve == nil || !approve(tc.Name, repairedArgs) {
				// Fail closed (N1): a RequireApproval verdict with no approver
				// configured denies the call instead of silently passing.
				st.Err = fmt.Sprintf("execution of tool %q denied: policy gate requires approval but no approver approved it", tc.Name)
				st.Result = st.Err
				return st
			} else {
				approved = true
			}
		case policy.Allow:
			// Allowed to proceed
		}
	}

	if mode == ModeReadOnly && tools.IsWriteTool(tc.Name) {
		st.Err = "blocked in read-only (ro) mode: switch to edit or bypass mode (Shift+Tab) to allow mutations"
		st.Result = st.Err
		return st
	}
	if mode == ModeEdit && !approved && tools.IsWriteTool(tc.Name) {
		// Fail closed (N1): edit mode requires approval for write tools; with
		// no approver configured the call is denied, not silently passed.
		if approve == nil || !approve(tc.Name, repairedArgs) {
			st.Err = "tool execution rejected: edit mode requires user approval for write tools and no approver approved it"
			st.Result = st.Err
			return st
		}
	}

	a.mu.RLock()
	hooks := a.cfg.Hooks
	a.mu.RUnlock()

	if hooks != nil {
		hCtx := &HookContext{
			Type:     HookPreTool,
			ToolName: tc.Name,
			ToolArgs: repairedArgs,
		}
		_ = hooks.Execute(ctx, hCtx)
	}

	tool, ok := reg.Get(tc.Name)
	if !ok {
		st.Err = "unknown tool: " + tc.Name
		st.Result = st.Err
		return st
	}
	r, err := tool.Run(ctx, args)
	st.Result = r
	if err != nil {
		st.Err = err.Error()
		if st.Result == "" {
			st.Result = st.Err
		}
	}

	if hooks != nil {
		hCtx := &HookContext{
			Type:       HookPostTool,
			ToolName:   tc.Name,
			ToolArgs:   repairedArgs,
			ToolResult: st.Result,
			ToolErr:    st.Err,
		}
		_ = hooks.Execute(ctx, hCtx)
		st.Result = hCtx.ToolResult
		if hCtx.ToolErr != "" {
			st.Err = hCtx.ToolErr
		}
	}
	return st
}

// buildPrompt selects smolSystemPrompt when SmolMode is active, or the standard
// systemPrompt otherwise, and appends the persistent harness-owned Todo checklist.
func (a *Agent) buildPrompt(memories []cortex.Neuron, schemas []llm.ToolSchema) string {
	return a.buildPromptForInput("", memories, schemas)
}

func (a *Agent) buildPromptForInput(input string, memories []cortex.Neuron, schemas []llm.ToolSchema) string {
	smol := a.SmolMode()
	base := systemPrompt(memories, schemas)
	if smol {
		base = smolSystemPrompt(memories, schemas)
	}
	return base + BuildPhaseScopedRulesBlock(a.SDLC(), input) + FormatTodoChecklist(a.todoLines())
}

// runLoop drives the infer -> act cycle until the model returns text or the step
// cap is hit. It returns the final text reply and a trace of tool executions.
func (a *Agent) runLoop(ctx context.Context, msgs []llm.Message) (string, []Step, error) {
	schemas := a.schemas()
	var trace []Step
	var last string
	smol := a.SmolMode()
	keepRecent := 4
	if smol {
		keepRecent = 2
	}
	for step := 0; step < maxToolSteps; step++ {
		msgs = a.maybeCompactContext(msgs, smol)
		pruned := PruneHistoryPlaceholders(msgs, keepRecent)
		a.mu.RLock()
		prov := a.cfg.Provider
		model := a.cfg.Model
		local := a.cfg.Local
		effort := a.cfg.Effort
		usageTracker := a.cfg.Usage
		a.mu.RUnlock()

		// In SmolMode, automatically disable chain-of-thought on mechanical follow-up
		// tool steps (step > 0) when effort is auto ("") for 3-5x faster tool loops.
		if smol && step > 0 && effort == "" {
			effort = "off"
		}

		resp, err := prov.Generate(ctx, llm.GenerateRequest{
			Model: model, Messages: pruned, Tools: schemas, Effort: effort,
		})
		if err != nil {
			return "", trace, err
		}
		if usageTracker != nil {
			usageTracker.Record(model, local, resp.Usage)
		}
		last = StripEmDashes(resp.Content)
		msgs = append(msgs, llm.Message{Role: "assistant", Content: last, ToolCalls: resp.ToolCalls})
		if len(resp.ToolCalls) == 0 {
			a.mu.RLock()
			hooks := a.cfg.Hooks
			a.mu.RUnlock()
			if hooks != nil {
				hCtx := &HookContext{
					Type:      HookOnStop,
					Trace:     trace,
					TodoLines: a.todoLines(),
				}
				_ = hooks.Execute(ctx, hCtx)
				if hCtx.InjectedMsg != "" {
					last += "\n\n[" + hCtx.InjectedMsg + "]"
				}
			}
			return last, trace, nil
		}
		for _, tc := range resp.ToolCalls {
			st := a.executeToolCall(ctx, tc)
			trace = append(trace, st)
			msgs = append(msgs, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: st.Result})
		}
		// Inject any mid-turn user steering messages queued during tool execution.
		for _, sm := range a.drainSteer() {
			msgs = append(msgs, llm.Message{Role: "user", Content: "[User steering instruction]: " + sm})
		}
	}
	a.mu.RLock()
	hooks := a.cfg.Hooks
	a.mu.RUnlock()
	if hooks != nil {
		hCtx := &HookContext{
			Type:      HookOnStop,
			Trace:     trace,
			TodoLines: a.todoLines(),
		}
		_ = hooks.Execute(ctx, hCtx)
		if hCtx.InjectedMsg != "" {
			last += "\n\n[" + hCtx.InjectedMsg + "]"
		}
	}
	return last, trace, nil
}

// Run executes one conversational turn and returns the final text reply.
func (a *Agent) Run(ctx context.Context, input string, history []llm.Message) (string, error) {
	r, err := a.runTurn(ctx, input, history)
	return r.Reply, err
}

// RunTrace is like Run but also returns the trace of tool executions.
func (a *Agent) RunTrace(ctx context.Context, input string, history []llm.Message) (string, []Step, error) {
	r, err := a.runTurn(ctx, input, history)
	return r.Reply, r.Trace, err
}

// RunFull returns the complete Result, including the memories recalled into the
// prompt - used by the trace CLI and the cross-run memory tests.
func (a *Agent) RunFull(ctx context.Context, input string, history []llm.Message) (Result, error) {
	return a.runTurn(ctx, input, history)
}

// RunStream runs one conversational turn while streaming updates to emit as they
// happen: recalled memories, token deltas of the model's reply, tool executions, and
// a final done event carrying the full reply (F2.1 streaming + tool calling).
func (a *Agent) RunStream(ctx context.Context, input string, history []llm.Message, emit func(Event)) error {
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		emit(Event{Type: EventError, Err: err.Error()})
		return err
	}
	for _, m := range memories {
		emit(Event{Type: EventMemory, Text: m.Content})
	}
	schemas := a.schemas()
	msgs := []llm.Message{
		{Role: "system", Content: a.buildPromptForInput(input, memories, schemas)},
	}
	msgs = append(msgs, history...)
	msgs = append(msgs, llm.Message{Role: "user", Content: input})
	var finalReply string
	smol := a.SmolMode()
	keepRecent := 4
	if smol {
		keepRecent = 2
	}
	for step := 0; step < maxToolSteps; step++ {
		msgs = a.maybeCompactContext(msgs, smol)
		pruned := PruneHistoryPlaceholders(msgs, keepRecent)
		a.mu.RLock()
		prov := a.cfg.Provider
		model := a.cfg.Model
		local := a.cfg.Local
		effort := a.cfg.Effort
		usageTracker := a.cfg.Usage
		a.mu.RUnlock()

		if smol && step > 0 && effort == "" {
			effort = "off"
		}

		ch, err := prov.Stream(ctx, llm.GenerateRequest{
			Model: model, Messages: pruned, Tools: schemas, Effort: effort,
		})
		if err != nil {
			emit(Event{Type: EventError, Err: err.Error()})
			return err
		}
		var content strings.Builder
		var toolCalls []llm.ToolCall
		var usage llm.Usage
		for ev := range ch {
			if ev.Err != nil {
				emit(Event{Type: EventError, Err: ev.Err.Error()})
				return ev.Err
			}
			if ev.Delta != "" {
				cleanDelta := StripEmDashes(ev.Delta)
				content.WriteString(cleanDelta)
				emit(Event{Type: EventToken, Text: cleanDelta})
			}
			if ev.Done {
				toolCalls = ev.ToolCalls
				usage = ev.Usage
			}
		}
		if usageTracker != nil {
			usageTracker.Record(model, local, usage)
		}
		text := StripEmDashes(content.String())
		finalReply = text
		msgs = append(msgs, llm.Message{Role: "assistant", Content: text, ToolCalls: toolCalls})
		if len(toolCalls) == 0 {
			a.mu.RLock()
			hooks := a.cfg.Hooks
			a.mu.RUnlock()
			if hooks != nil {
				hCtx := &HookContext{
					Type:      HookOnStop,
					TodoLines: a.todoLines(),
				}
				_ = hooks.Execute(ctx, hCtx)
				if hCtx.InjectedMsg != "" {
					finalReply += "\n\n[" + hCtx.InjectedMsg + "]"
				}
			}
			emit(Event{Type: EventDone, Reply: finalReply})
			return nil
		}
		for _, tc := range toolCalls {
			st := a.executeToolCall(ctx, tc)
			emit(Event{Type: EventTool, Step: &st})
			msgs = append(msgs, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: st.Result})
		}
		for _, sm := range a.drainSteer() {
			emit(Event{Type: EventSteer, Text: sm})
			msgs = append(msgs, llm.Message{Role: "user", Content: "[User steering instruction]: " + sm})
		}
	}
	a.mu.RLock()
	hooks := a.cfg.Hooks
	a.mu.RUnlock()
	if hooks != nil {
		hCtx := &HookContext{
			Type:      HookOnStop,
			TodoLines: a.todoLines(),
		}
		_ = hooks.Execute(ctx, hCtx)
		if hCtx.InjectedMsg != "" {
			finalReply += "\n\n[" + hCtx.InjectedMsg + "]"
		}
	}
	emit(Event{Type: EventDone, Reply: finalReply})
	return nil
}

func (a *Agent) runTurn(ctx context.Context, input string, history []llm.Message) (Result, error) {
	memories, err := a.cfg.Cortex.Search(ctx, input, 5)
	if err != nil {
		return Result{}, err
	}
	recalled := make([]string, len(memories))
	for i, m := range memories {
		recalled[i] = m.Content
	}
	schemas := a.schemas()
	msgs := []llm.Message{
		{Role: "system", Content: a.buildPromptForInput(input, memories, schemas)},
	}
	msgs = append(msgs, history...)
	msgs = append(msgs, llm.Message{Role: "user", Content: input})
	reply, trace, err := a.runLoop(ctx, msgs)
	if err != nil {
		return Result{Trace: trace, Recalled: recalled}, err
	}
	return Result{Reply: reply, Trace: trace, Recalled: recalled}, nil
}

// systemPrompt assembles the system prompt with recalled memory + a tool hint.
func systemPrompt(memories []cortex.Neuron, schemas []llm.ToolSchema) string {
	p := "You are Mímir, an agent that remembers. Use what you recall. You have access to the full conversation history - reference earlier messages when relevant."
	if len(memories) > 0 {
		p += "\n\nRelevant memory:"
		for _, m := range memories {
			p += "\n- " + m.Content
		}
	}
	if len(schemas) > 0 {
		names := make([]string, len(schemas))
		for i, s := range schemas {
			names[i] = s.Name
		}
		p += "\n\nYou have these tools - call them when useful: " + strings.Join(names, ", ") + "."
	}
	p += "\n\nWhen presenting multiple-choice questions or options, use this exact format so the UI renders clickable buttons:\n[choice]\nA) option text\nB) option text\nC) option text\n[/choice]\nAlways use this format for multiple-choice answers - never just list them as plain text."
	return p
}

// maybeCompactContext checks if estimated conversation token count exceeds 80%
// of the context budget. When exceeded, it generates `.mimir/HANDOFF.md` to persist
// continuity state before aggressively compacting history to prevent context overflow.
func (a *Agent) maybeCompactContext(msgs []llm.Message, smol bool) []llm.Message {
	budget := a.cfg.ContextBudget
	if budget <= 0 {
		if smol {
			budget = 8192
		} else {
			budget = 32768
		}
	}
	threshold := int(float64(budget) * 0.8)
	current := EstimateTokens(msgs)
	if current < threshold {
		return msgs
	}

	// Generate continuity handoff document before aggressive trimming
	handoffMD := GenerateHandoffMarkdown("session-auto-compaction", msgs, a.todoLines(), "")
	_, _ = WriteHandoffFile(".", handoffMD)

	// Keep system prompt (index 0) and the most recent 2 turns, replace middle with summary notice
	if len(msgs) <= 3 {
		return msgs
	}
	sysMsg := msgs[0]
	recent := msgs[len(msgs)-2:]
	compactedNotice := llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("[Context Compaction: Conversation exceeded 80%% context threshold (%d/%d tokens). Previous state persisted to .mimir/HANDOFF.md. Active todos and recent turns preserved.]", current, budget),
	}
	return append([]llm.Message{sysMsg, compactedNotice}, recent...)
}
