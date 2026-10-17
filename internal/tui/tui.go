package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/config"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/ingest"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// Options configures the interactive Mímir TUI session.
type Options struct {
	Version    string
	Agent      *agent.Agent
	Tools      *tools.Registry
	Cortex     cortex.Store
	Sessions   cortex.SessionStore
	Usage      *llm.Tracker
	StoreDesc  string
	WebURL     string
	StartWebFn func() (string, error) // optional callback to start the Book-Spine Web UI on /web
	In         io.Reader
	Out        io.Writer
}

// Run launches the interactive Mímir terminal UI loop.
func Run(ctx context.Context, opts Options) error {
	in := opts.In
	if in == nil {
		in = os.Stdin
	}
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	// Wire QuestionTool in the registry to our interactive terminal selector.
	if opts.Tools != nil {
		if qt, ok := opts.Tools.Get("question"); ok {
			if qTool, ok := qt.(*tools.QuestionTool); ok {
				qTool.AskFunc = func(qctx context.Context, qs []tools.Question) (string, error) {
					return PromptQuestions(qctx, in, out, qs)
				}
			}
		}
	}

	pid, model, local := opts.Agent.ModelInfo()
	toolCount := 0
	if opts.Tools != nil {
		toolCount = len(opts.Tools.ForMode(opts.Agent.Mode()))
	}
	fmt.Fprint(out, RenderBanner(
		opts.Version, pid, model, opts.Agent.Mode(), opts.Agent.Effort(),
		opts.StoreDesc, opts.WebURL, toolCount,
	))

	var history []llm.Message
	var activePersona *agent.Persona
	scanner := bufio.NewScanner(in)
	// Allow up to 1MB pasted prompts in the terminal.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for {
		pid, model, local = opts.Agent.ModelInfo()
		totalTok := 0
		if opts.Usage != nil {
			totalTok = opts.Usage.TotalTokens()
		}
		estTok := agent.EstimateTokens(history)
		fmt.Fprintln(out, RenderStatusBar(pid, model, opts.Agent.Mode(), opts.Agent.Effort(), estTok, totalTok, local, opts.Agent.SmolMode()))

		promptLabel := "mimir"
		if activePersona != nil {
			promptLabel = "mimir:" + activePersona.Name
		}
		fmt.Fprintf(out, "%s%s ❯%s ", ansiGold+ansiBold, promptLabel, ansiReset)

		if !scanner.Scan() {
			fmt.Fprintln(out)
			return nil
		}
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "/") {
			quit, err := handleSlashCommand(ctx, line, &opts, &history, &activePersona, scanner, out)
			if err != nil {
				fmt.Fprintf(out, "%s  error: %v%s\n", ansiRed, err, ansiReset)
			}
			if quit {
				return nil
			}
			continue
		}

		if pid == "none" || model == "" {
			fmt.Fprintf(out, "%s  No model configured. Run /model to auto-detect local Ollama/LM Studio or set MIMIR_BASE_URL.%s\n\n", ansiRed, ansiReset)
			continue
		}

		userPrompt := line
		if activePersona != nil {
			userPrompt = fmt.Sprintf("[Persona: %s (%s)]\n%s\n\nUser request: %s",
				activePersona.Name, activePersona.Role, activePersona.FullPrompt(), line)
		}

		var replyBuf strings.Builder
		var streamedTokens bool
		fmt.Fprintln(out)

		err := opts.Agent.RunStream(ctx, userPrompt, history, func(ev agent.Event) {
			switch ev.Type {
			case agent.EventMemory:
				fmt.Fprintf(out, "  %s◈ recalled:%s %s%s%s\n",
					ansiPurple, ansiReset, ansiDim, truncate(ev.Text, 85), ansiReset)
			case agent.EventTool:
				if streamedTokens {
					fmt.Fprintln(out)
					streamedTokens = false
				}
				if ev.Step != nil {
					fmt.Fprint(out, RenderToolPill(ev.Step.Name, ev.Step.Args, ev.Step.Result, ev.Step.Err))
				}
			case agent.EventToken:
				if !streamedTokens {
					fmt.Fprint(out, "  ")
					streamedTokens = true
				}
				replyBuf.WriteString(ev.Text)
				fmt.Fprint(out, strings.ReplaceAll(ev.Text, "\n", "\n  "))
			case agent.EventSteer:
				fmt.Fprintf(out, "\n  %s⚡ steered:%s %s\n", ansiGold, ansiReset, ev.Text)
			case agent.EventError:
				fmt.Fprintf(out, "\n  %s✗ %s%s\n", ansiRed, ev.Err, ansiReset)
			case agent.EventDone:
				if !streamedTokens && ev.Reply != "" {
					fmt.Fprintln(out, "  "+strings.ReplaceAll(FormatMarkdown(ev.Reply), "\n", "\n  "))
				} else if streamedTokens {
					fmt.Fprintln(out)
				}
			}
		})
		if err != nil {
			fmt.Fprintf(out, "  %serror: %v%s\n\n", ansiRed, err, ansiReset)
			continue
		}
		fmt.Fprintln(out)

		finalReply := strings.TrimSpace(replyBuf.String())
		history = append(history,
			llm.Message{Role: "user", Content: line},
			llm.Message{Role: "assistant", Content: finalReply},
		)
	}
}

func handleSlashCommand(
	ctx context.Context,
	line string,
	opts *Options,
	history *[]llm.Message,
	activePersona **agent.Persona,
	scanner *bufio.Scanner,
	out io.Writer,
) (bool, error) {
	fields := strings.Fields(line)
	cmd := strings.ToLower(fields[0])
	arg := ""
	if len(fields) > 1 {
		arg = strings.TrimSpace(strings.Join(fields[1:], " "))
	}

	switch cmd {
	case "/quit", "/exit", "/q":
		fmt.Fprintf(out, "%s  Until next time. Mímir remembers.%s\n", ansiDim, ansiReset)
		return true, nil

	case "/help", "/?":
		fmt.Fprintf(out, `
%s  Mímir Slash Commands:%s
    %s/model [name]%s              Auto-discover local Ollama/LM Studio models or switch model
    %s/providers%s                 Show built-in BYOK providers + Mímir Well Subscription tiers
    %s/ingest [path|git-url]%s     GitIngest a repo/folder into digest + Yggdrasil Mind Map in Cortex
    %s/mindmap [query]%s           View the Yggdrasil Mind Map or ask Liquid d1-omni-600M for top files
    %s/mode [ro|edit|bypass]%s     Cycle or set permission mode (ro hides write tools from model!)
    %s/effort [off|low|med|high]%s Toggle reasoning/thinking budget (off = fast /no_think loops)
    %s/smol [auto|on|off]%s        Toggle SmolMode optimizations for compact/local models
    %s/decide [prompt]%s           Run Liquid AI d1-omni-600M System-One decision check on a prompt
    %s/persona [name]%s            Switch Norse persona (odin, thor, loki, heimdall, bragi, huginn, muninn, forseti)
    %s/todo%s                      Show the active harness-owned Todo checklist
    %s/memory%s                    Show stored Cortex memories
    %s/compact%s                   Prune old file-write payloads from conversation history now
    %s/web%s                       Start or show the live Book-Spine Web UI URL
    %s/clear%s                     Clear current conversation history
    %s/quit%s                      Exit Mímir TUI
`+"\n", ansiGold+ansiBold, ansiReset,
			ansiCyan, ansiReset, ansiCyan, ansiReset, ansiCyan, ansiReset,
			ansiCyan, ansiReset, ansiCyan, ansiReset, ansiCyan, ansiReset,
			ansiCyan, ansiReset, ansiCyan, ansiReset, ansiCyan, ansiReset,
			ansiCyan, ansiReset, ansiCyan, ansiReset, ansiCyan, ansiReset,
			ansiCyan, ansiReset, ansiCyan, ansiReset, ansiCyan, ansiReset)
		return false, nil

	case "/mode":
		var next string
		if arg == "" {
			next = opts.Agent.CycleMode()
		} else {
			switch strings.ToLower(arg) {
			case "ro", "readonly", "read-only":
				opts.Agent.SetMode(agent.ModeReadOnly)
			case "edit":
				opts.Agent.SetMode(agent.ModeEdit)
			case "bypass", "yolo":
				opts.Agent.SetMode(agent.ModeBypass)
			default:
				return false, fmt.Errorf("unknown mode %q (use ro, edit, or bypass)", arg)
			}
			next = opts.Agent.Mode()
		}
		toolCount := 0
		if opts.Tools != nil {
			toolCount = len(opts.Tools.ForMode(next))
		}
		fmt.Fprintf(out, "  %s✓ Mode -> %s%s (%d tools visible to model)\n\n", ansiGreen, next, ansiReset, toolCount)
		return false, nil

	case "/effort":
		switch strings.ToLower(arg) {
		case "off", "none":
			opts.Agent.SetEffort("off")
		case "low":
			opts.Agent.SetEffort("low")
		case "med", "medium":
			opts.Agent.SetEffort("medium")
		case "high":
			opts.Agent.SetEffort("high")
		case "auto", "default", "":
			if arg == "" {
				// Cycle: "" -> "off" -> "low" -> "medium" -> "high" -> ""
				switch opts.Agent.Effort() {
				case "":
					opts.Agent.SetEffort("off")
				case "off":
					opts.Agent.SetEffort("low")
				case "low":
					opts.Agent.SetEffort("medium")
				case "medium":
					opts.Agent.SetEffort("high")
				default:
					opts.Agent.SetEffort("")
				}
			} else {
				opts.Agent.SetEffort("")
			}
		default:
			return false, fmt.Errorf("unknown effort %q (use off, low, medium, high, or auto)", arg)
		}
		eff := opts.Agent.Effort()
		if eff == "" {
			eff = "auto"
		}
		fmt.Fprintf(out, "  %s✓ Reasoning effort -> %s%s\n\n", ansiGreen, eff, ansiReset)
		return false, nil

	case "/smol":
		switch strings.ToLower(arg) {
		case "on", "true", "1":
			opts.Agent.SetSmolPolicy(agent.SmolOn)
		case "off", "false", "0":
			opts.Agent.SetSmolPolicy(agent.SmolOff)
		default:
			opts.Agent.SetSmolPolicy(agent.SmolAuto)
		}
		fmt.Fprintf(out, "  %s✓ SmolMode policy -> %s (active=%v)%s\n\n",
			ansiGreen, opts.Agent.SmolPolicy(), opts.Agent.SmolMode(), ansiReset)
		return false, nil

	case "/decide":
		dec := opts.Agent.Decider()
		if arg == "" {
			fmt.Fprintf(out, "  %s⚡ System-One Decision Engine: %s%s\n", ansiGold+ansiBold, dec.Model, ansiReset)
			fmt.Fprintf(out, "  %sUsage: /decide <prompt or command> to inspect routing, tool pruning, Mind Map files, and safety score.%s\n\n", ansiDim, ansiReset)
			return false, nil
		}
		route := dec.RoutePrompt(ctx, arg)
		safety := dec.AssessCommandSafety(ctx, arg)
		var toolNames []string
		if opts.Tools != nil {
			for _, t := range opts.Tools.ForMode(opts.Agent.Mode()) {
				toolNames = append(toolNames, t.Name())
			}
		}
		pruned := dec.PruneToolNames(ctx, arg, toolNames, 5)
		fmt.Fprintf(out, "\n  %s⚡ Huginn-Decider (%s):%s\n", ansiGold+ansiBold, dec.Model, ansiReset)
		fmt.Fprintf(out, "    %s• Route:%s Persona=%s (%s tier, thinking=%v, conf=%.0f%%)\n",
			ansiCyan, ansiReset, route.Persona, route.Tier, route.NeedsThinking, route.Confidence*100)
		fmt.Fprintf(out, "    %s• Pruned Tools (top %d):%s %s\n",
			ansiCyan, len(pruned), ansiReset, strings.Join(pruned, ", "))
		if mm, ok := ingest.LatestMindMap(opts.Cortex); ok {
			var entries []string
			for _, n := range mm.Nodes {
				entries = append(entries, n.Path+" "+strings.Join(n.Symbols, " "))
			}
			topFiles := dec.SelectMindMapFiles(ctx, arg, entries, 4)
			fmt.Fprintf(out, "    %s• Yggdrasil Mind Map (%s):%s %s\n",
				ansiCyan, mm.RepoName, ansiReset, strings.Join(topFiles, " | "))
		}
		fmt.Fprintf(out, "    %s• Safety Gate:%s score=%.2f (auto_approve=%v, %s)\n\n",
			ansiCyan, ansiReset, safety.SafeScore, safety.AutoApprove, safety.Reason)
		return false, nil

	case "/ingest", "/gitingest":
		src := arg
		if src == "" {
			src = "."
		}
		fmt.Fprintf(out, "  %s◈ Gjallarhorn-Ingest scanning %s...%s\n", ansiPurple, src, ansiReset)
		res, err := ingest.Run(ctx, ingest.Options{Source: src}, opts.Cortex)
		if err != nil {
			return false, err
		}
		fmt.Fprintf(out, "  %s✓ Ingested %s (%d files, ~%dk tokens) -> %d Yggdrasil Mind Map neurons saved to Cortex%s\n\n",
			ansiGreen, res.MindMap.RepoName, res.MindMap.TotalFiles, res.MindMap.EstTokens/1000, res.NeuronsSaved, ansiReset)
		return false, nil

	case "/mindmap", "/map":
		mm, ok := ingest.LatestMindMap(opts.Cortex)
		if !ok {
			fmt.Fprintf(out, "  %sNo Yggdrasil Mind Map in Cortex yet. Run `/ingest .` or `/ingest <git-url>` first.%s\n\n", ansiDim, ansiReset)
			return false, nil
		}
		if arg == "" {
			fmt.Fprintln(out, "\n"+ingest.FormatMindMapCompact(*mm, 25))
			return false, nil
		}
		dec := opts.Agent.Decider()
		var entries []string
		var testFiles []string
		for _, n := range mm.Nodes {
			entries = append(entries, n.Path+" ("+strings.Join(n.Symbols, ",")+")")
			if n.IsTest {
				testFiles = append(testFiles, n.Path)
			}
		}
		top := dec.SelectMindMapFiles(ctx, arg, entries, 5)
		tests := dec.SelectAffectedTests(ctx, top, testFiles, 3)
		fmt.Fprintf(out, "\n  %s🌳 Yggdrasil Mind Map + Liquid AI d1-omni-600M (%s):%s\n", ansiGold+ansiBold, mm.RepoName, ansiReset)
		for _, f := range top {
			fmt.Fprintf(out, "    %s• File:%s %s\n", ansiCyan, ansiReset, f)
		}
		if len(tests) > 0 {
			fmt.Fprintf(out, "    %s• Affected Tests:%s %s\n", ansiPurple, ansiReset, strings.Join(tests, ", "))
		}
		fmt.Fprintln(out)
		return false, nil

	case "/providers", "/subscribe":
		fmt.Fprintf(out, "\n%s  ╭─ Mímir Well Subscriptions & BYOK Provider Catalog ─────────────────╮%s\n", ansiGold+ansiBold, ansiReset)
		for _, sub := range config.BuiltinSubscriptionTiers() {
			fmt.Fprintf(out, "  %s│%s %s%s (%s)%s - %s\n",
				ansiGold, ansiReset, ansiCyan+ansiBold, sub.Name, sub.Price, ansiReset, sub.Description)
		}
		fmt.Fprintf(out, "  %s├────────────────────────────────────────────────────────────────────┤%s\n", ansiGold, ansiReset)
		for _, p := range config.BuiltinProviders() {
			badge := ansiDim + "[BYOK]" + ansiReset
			if p.BillingMode == "subscription" {
				badge = ansiGold + "[SUB]" + ansiReset
			} else if p.BillingMode == "local_free" {
				badge = ansiGreen + "[LOCAL]" + ansiReset
			}
			status := ansiDim + "unset" + ansiReset
			if p.Configured {
				status = ansiGreen + "ready" + ansiReset
			}
			fmt.Fprintf(out, "  %s│%s %s %-28s default=%-22s (%s)\n",
				ansiGold, ansiReset, badge, p.Name, p.DefaultModel, status)
		}
		fmt.Fprintf(out, "%s  ╰────────────────────────────────────────────────────────────────────╯%s\n\n", ansiGold+ansiBold, ansiReset)
		return false, nil

	case "/sdlc", "/phase":
		mgr := opts.Agent.SDLC()
		sub := strings.ToLower(strings.TrimSpace(arg))
		switch sub {
		case "":
			fmt.Fprintln(out, "\n"+mgr.FormatBoard()+"\n")
			return false, nil
		case "next", "advance":
			spec, err := mgr.Advance()
			if err != nil {
				return false, err
			}
			fmt.Fprintf(out, "  %s✓ SDLC advanced -> [%d/7] %s (Persona: %s | Deliverable: %s)%s\n\n",
				ansiGreen, spec.Step, spec.Title, spec.Persona, spec.Deliverable, ansiReset)
			return false, nil
		case "signoff", "sign-off", "sign":
			spec := mgr.SignOff("")
			fmt.Fprintf(out, "  %s✓ Signed off SDLC Step %d (%s - %s)%s\n\n",
				ansiGreen, spec.Step, spec.Title, spec.Deliverable, ansiReset)
			return false, nil
		case "agile", "waterfall", "devops":
			if err := mgr.SetMethodology(sub); err != nil {
				return false, err
			}
			fmt.Fprintf(out, "  %s✓ SDLC methodology -> %s%s\n\n", ansiGreen, mgr.Methodology(), ansiReset)
			return false, nil
		default:
			spec, err := mgr.SetPhase(sub)
			if err != nil {
				return false, err
			}
			fmt.Fprintf(out, "  %s✓ SDLC phase -> [%d/7] %s (Persona: %s | Deliverable: %s)%s\n\n",
				ansiGreen, spec.Step, spec.Title, spec.Persona, spec.Deliverable, ansiReset)
			return false, nil
		}

	case "/model", "/models":
		discovered := config.DiscoverLocalModels(ctx, 500*time.Millisecond)
		if arg != "" {
			// Direct switch by index or model name
			if idx, err := strconv.Atoi(arg); err == nil && idx >= 1 && idx <= len(discovered) {
				m := discovered[idx-1]
				prov := &llm.OpenAIProvider{IDStr: m.ProviderID, BaseURL: m.BaseURL, APIKey: m.APIKey, Model: m.Model}
				opts.Agent.SetProvider(prov, m.Model, m.Local)
				fmt.Fprintf(out, "  %s✓ Switched to %s/%s%s\n\n", ansiGreen, m.ProviderID, m.Model, ansiReset)
				return false, nil
			}
			for _, m := range discovered {
				if strings.EqualFold(m.Model, arg) {
					prov := &llm.OpenAIProvider{IDStr: m.ProviderID, BaseURL: m.BaseURL, APIKey: m.APIKey, Model: m.Model}
					opts.Agent.SetProvider(prov, m.Model, m.Local)
					fmt.Fprintf(out, "  %s✓ Switched to %s/%s%s\n\n", ansiGreen, m.ProviderID, m.Model, ansiReset)
					return false, nil
				}
			}
			// Keep current provider base URL and swap model name
			if ep, ok := config.AutoDiscoverProvider(ctx); ok {
				prov := &llm.OpenAIProvider{IDStr: ep.ID, BaseURL: ep.BaseURL, APIKey: ep.APIKey, Model: arg}
				local := strings.Contains(ep.BaseURL, "localhost") || strings.Contains(ep.BaseURL, "127.0.0.1")
				opts.Agent.SetProvider(prov, arg, local)
				fmt.Fprintf(out, "  %s✓ Switched model to %s (%s)%s\n\n", ansiGreen, arg, ep.ID, ansiReset)
				return false, nil
			}
		}

		if len(discovered) == 0 {
			fmt.Fprintf(out, "  %sNo local Ollama (:11434) or LM Studio (:1234) models detected.%s\n", ansiDim, ansiReset)
			fmt.Fprintf(out, "  %sTip: run `/model <model-id>` to switch models on your configured provider.%s\n\n", ansiDim, ansiReset)
			return false, nil
		}
		fmt.Fprintf(out, "\n%s  Discovered Local Models:%s\n", ansiGold+ansiBold, ansiReset)
		for i, m := range discovered {
			fmt.Fprintf(out, "    %s[%d]%s %s/%s %s(%s)%s\n",
				ansiCyan+ansiBold, i+1, ansiReset, m.ProviderID, m.Model, ansiDim, m.BaseURL, ansiReset)
		}
		fmt.Fprintf(out, "  %sSelect number (1-%d) or Enter to cancel:%s ", ansiDim, len(discovered), ansiReset)
		if scanner.Scan() {
			choice := strings.TrimSpace(scanner.Text())
			if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(discovered) {
				m := discovered[n-1]
				prov := &llm.OpenAIProvider{IDStr: m.ProviderID, BaseURL: m.BaseURL, APIKey: m.APIKey, Model: m.Model}
				opts.Agent.SetProvider(prov, m.Model, m.Local)
				fmt.Fprintf(out, "  %s✓ Switched to %s/%s%s\n\n", ansiGreen, m.ProviderID, m.Model, ansiReset)
			}
		}
		return false, nil

	case "/todo", "/checklist":
		if opts.Tools == nil {
			return false, nil
		}
		t, ok := opts.Tools.Get("todo")
		if !ok {
			return false, nil
		}
		tt, ok := t.(*tools.TodoTool)
		if !ok || tt.Store == nil {
			return false, nil
		}
		items := tt.Store.List()
		if len(items) == 0 {
			fmt.Fprintf(out, "  %sTodo checklist is empty.%s\n\n", ansiDim, ansiReset)
			return false, nil
		}
		fmt.Fprintf(out, "\n%s  ╭─ Harness Todo Checklist ───────────────────────────╮%s\n", ansiGold, ansiReset)
		for _, item := range items {
			badge := "[ ]"
			col := ansiDim
			switch item.Status {
			case tools.TodoCompleted:
				badge = "[✓]"
				col = ansiGreen
			case tools.TodoInProgress:
				badge = "[▸]"
				col = ansiCyan
			case tools.TodoBlocked:
				badge = "[!]"
				col = ansiRed
			}
			fmt.Fprintf(out, "  %s│%s %s%s #%d %s%s\n", ansiGold, ansiReset, col, badge, item.ID, item.Content, ansiReset)
		}
		fmt.Fprintf(out, "%s  ╰────────────────────────────────────────────────────╯%s\n\n", ansiGold, ansiReset)
		return false, nil

	case "/memory", "/cortex":
		if opts.Cortex == nil {
			return false, nil
		}
		neurons := opts.Cortex.All()
		if len(neurons) == 0 {
			fmt.Fprintf(out, "  %sNo neurons stored in Cortex yet.%s\n\n", ansiDim, ansiReset)
			return false, nil
		}
		fmt.Fprintf(out, "\n%s  Cortex Memories (%d):%s\n", ansiPurple+ansiBold, len(neurons), ansiReset)
		for i, n := range neurons {
			if i >= 15 {
				fmt.Fprintf(out, "    %s... and %d more%s\n", ansiDim, len(neurons)-15, ansiReset)
				break
			}
			fmt.Fprintf(out, "    %s• [%s]%s %s\n", ansiPurple, n.Kind, ansiReset, truncate(n.Content, 80))
		}
		fmt.Fprintln(out)
		return false, nil

	case "/persona":
		if arg == "" || arg == "clear" || arg == "off" {
			*activePersona = nil
			fmt.Fprintf(out, "  %sCleared persona (using default Mímir voice).%s\n", ansiDim, ansiReset)
			var names []string
			for _, p := range agent.BuiltinPersonas {
				names = append(names, fmt.Sprintf("%s (%s)", p.Name, p.Role))
			}
			fmt.Fprintf(out, "  %sAvailable: %s%s\n\n", ansiDim, strings.Join(names, ", "), ansiReset)
			return false, nil
		}
		p, ok := agent.PersonaByName(strings.ToLower(arg))
		if !ok {
			return false, fmt.Errorf("unknown persona %q", arg)
		}
		*activePersona = &p
		fmt.Fprintf(out, "  %s✓ Persona active: %s (%s)%s\n\n", ansiGreen, p.Name, p.Role, ansiReset)
		return false, nil

	case "/compact":
		before := agent.EstimateTokens(*history)
		*history = agent.PruneHistoryPlaceholders(*history, 2)
		after := agent.EstimateTokens(*history)
		fmt.Fprintf(out, "  %s✓ History compacted (%d -> %d est. tokens)%s\n\n", ansiGreen, before, after, ansiReset)
		return false, nil

	case "/web":
		if opts.WebURL != "" {
			fmt.Fprintf(out, "  %s● Book-Spine Web UI is live at %s%s\n\n", ansiGold, opts.WebURL, ansiReset)
			return false, nil
		}
		if opts.StartWebFn != nil {
			url, err := opts.StartWebFn()
			if err != nil {
				return false, err
			}
			opts.WebURL = url
			fmt.Fprintf(out, "  %s✓ Started Book-Spine Web UI at %s%s\n\n", ansiGreen, url, ansiReset)
			return false, nil
		}
		fmt.Fprintf(out, "  %sRun `mimir --web` or `mimir serve` to enable the Web UI.%s\n\n", ansiDim, ansiReset)
		return false, nil

	case "/clear":
		*history = nil
		fmt.Fprintf(out, "  %s✓ Conversation history cleared.%s\n\n", ansiGreen, ansiReset)
		return false, nil

	default:
		return false, fmt.Errorf("unknown command %q (type /help for commands)", cmd)
	}
}
