// Command mimir is the entrypoint for Mímir - the agent that remembers.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/config"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/critic"
	"github.com/David2024patton/Mimir/internal/decision"
	"github.com/David2024patton/Mimir/internal/engines"
	"github.com/David2024patton/Mimir/internal/ingest"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/runengine"
	"github.com/David2024patton/Mimir/internal/server"
	"github.com/David2024patton/Mimir/internal/tools"
	"github.com/David2024patton/Mimir/internal/tui"
)

const version = "0.2.0-dev"

const usage = `Mímir - the agent that remembers (tiny self-hosted CLI, TUI & Book-Spine Web UI)

Usage:
  mimir                       launch interactive terminal UI (auto-discovers Ollama/LM Studio/Cloud)
  mimir --web [--addr :8420]  launch interactive TUI + live Book-Spine Web UI simultaneously
  mimir ingest [path|git-url] GitIngest a repo/folder into digest.txt + Yggdrasil Mind Map in Cortex
  mimir sdlc [phase|method]   show or switch the 7-Phase SDLC board & methodology (agile/waterfall/devops)
  mimir review [draft] [ctx]  run the Tier-2 Deterministic Review Gate (Checklist + SDLC + Full Suite + Critic)
  mimir ship [draft] [ctx]    run the Tier-2 Production Ship Gate and advance SDLC to Deployment
  mimir audit <draft> [ctx]   run the Epistemic Critic & Verification Node against a draft report
  mimir audit-web [dir]       run Impeccable anti-slop UI + 4-Layer SEO/GEO & Conversion detectors
  mimir engines               show Alpha Flux LLC engine binaries, versions, and integrity
  mimir engines run <id> ...  run an engine binary (attribution handshake added automatically)
  mimir engines publish       (maintainers) rehash engines/bin/* and rewrite manifest.json
  mimir update [--check]      update the harness (git fast-forward + rebuild) and any changed engines
  mimir handoff               generate .mimir/HANDOFF.md for session continuity
  mimir providers             list built-in BYOK providers + Mímir Well Subscription tiers
  mimir setup                 bootstrap Liquid AI d1-omni-600M local System-One decision model
  mimir -p "your prompt"      headless single-shot prompt (for scripts & CI)
  mimir "your prompt"         run one turn (final reply only)
  mimir trace "your prompt"   run one turn and print recalled memory + every tool call
  mimir serve [--addr :8420]  serve the live book-spine UI + HTTP wire in foreground
  mimir status                show provider, discovered models, tools, and memory status
  mimir version               print version

Flags:
  -p, --prompt <text>         run a single prompt non-interactively and exit
  --web                       also serve the Book-Spine Web UI while in the TUI
  --addr <:port>              bind address for Web UI (default :8420)
  --mode <ro|edit|bypass>     initial permission mode (default: edit; ro hides write tools)
  --effort <off|low|med|high> reasoning effort budget (off = fast /no_think tool loops)
  --model <model-id>          override model id

Provider config (auto-discovered or environment):
  Auto-probed local servers:  Ollama (127.0.0.1:11434), LM Studio (127.0.0.1:1234)
  MIMIR_BASE_URL              OpenAI-compatible base URL, e.g. https://api.openai.com/v1
  MIMIR_API_KEY               bearer key (optional for local servers; also checks MIMO_API_KEY)
  MIMIR_MODEL                 model id (optional)
  MIMIR_PROVIDER_ID           provider id label (default: openai-compatible)

Memory (the Cortex) persists across runs:
  MIMIR_HOME                  memory dir (default: ./.mimir in the working directory)
  MIMIR_CORTEX                full path to the cortex file (default: <MIMIR_HOME>/cortex.json)
  MIMIR_CORTEX_BACKEND        file (default) | surreal`

// isLocalURL reports whether a provider base URL points at a local server (Ollama,
// vLLM, Mímir's own gateway). Local tokens cost $0 on the dashboard (F52).
func isLocalURL(u string) bool {
	return strings.Contains(u, "localhost") || strings.Contains(u, "127.0.0.1") || strings.Contains(u, "0.0.0.0")
}

// buildCortex creates the Cortex store for the configured backend.
func buildCortex(home string) (cortex.Store, cortex.SessionStore, *cortex.AuthStore, string, func(), error) {
	noop := func() {}
	if strings.EqualFold(os.Getenv("MIMIR_CORTEX_BACKEND"), "surreal") {
		surrealAddr := os.Getenv("MIMIR_SURREAL_SERVER")
		var sc *cortex.Sidecar
		ctx := context.Background()
		if surrealAddr == "" {
			addr := os.Getenv("MIMIR_SURREAL_ADDR")
			if addr == "" {
				addr = "127.0.0.1:8000"
			}
			dataPath := os.Getenv("MIMIR_SURREAL_DATA")
			if dataPath == "" {
				dataPath = filepath.Join(home, "surreal")
			}
			sc = &cortex.Sidecar{Addr: addr, DataPath: dataPath}
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 25*time.Second)
			defer cancel()
			if err := sc.Start(ctx); err != nil {
				return nil, nil, nil, "", noop, fmt.Errorf("surreal sidecar: %w", err)
			}
			surrealAddr = sc.HTTPAddr()
		}
		embedURL := os.Getenv("MIMIR_EMBED_URL")
		if embedURL == "" {
			embedURL = strings.TrimSuffix(strings.TrimRight(os.Getenv("MIMIR_BASE_URL"), "/"), "/v1")
		}
		if embedURL == "" {
			embedURL = "http://localhost:11434"
		}
		embedModel := os.Getenv("MIMIR_EMBED_MODEL")
		if embedModel == "" {
			embedModel = "nomic-embed-text"
		}
		dim := 768
		if d := os.Getenv("MIMIR_EMBED_DIM"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				dim = n
			}
		}
		store, err := cortex.NewSurrealStore(ctx, cortex.SurrealConfig{
			Addr:      surrealAddr,
			Dimension: dim,
			Embedder:  &cortex.OllamaEmbedder{BaseURL: embedURL, Model: embedModel},
		})
		if err != nil {
			if sc != nil {
				sc.Stop()
			}
			return nil, nil, nil, "", noop, fmt.Errorf("surreal store: %w", err)
		}
		cli := store.Client()
		sessStore := cortex.NewSessionStore(cli)
		if err := sessStore.EnsureSessionsTable(ctx); err != nil {
			if sc != nil {
				sc.Stop()
			}
			return nil, nil, nil, "", noop, fmt.Errorf("sessions table: %w", err)
		}
		authStore := cortex.NewAuthStore(cli)
		if err := authStore.EnsureAuthTables(ctx); err != nil {
			if sc != nil {
				sc.Stop()
			}
			return nil, nil, nil, "", noop, fmt.Errorf("auth tables: %w", err)
		}
		adminEmail := os.Getenv("MIMIR_ADMIN_EMAIL")
		if adminEmail == "" {
			adminEmail = "david@itak.live"
		}
		if err := authStore.SeedAdmin(ctx, adminEmail); err != nil {
			if sc != nil {
				sc.Stop()
			}
			return nil, nil, nil, "", noop, fmt.Errorf("seed admin: %w", err)
		}
		fmt.Fprintf(os.Stderr, "auth: super admin %s seeded\n", adminEmail)
		stop := noop
		if sc != nil {
			stop = sc.Stop
		}
		desc := fmt.Sprintf("SurrealDB %s (embed %s, dim %d)", surrealAddr, embedModel, dim)
		return store, sessStore, authStore, desc, stop, nil
	}
	cpath := os.Getenv("MIMIR_CORTEX")
	if cpath == "" {
		cpath = filepath.Join(home, "cortex.json")
	}
	store, err := cortex.NewMemoryStoreAt(cpath)
	if err != nil {
		return nil, nil, nil, "", noop, err
	}
	sessStore, err := cortex.NewFileSessionStore(home)
	if err != nil {
		return nil, nil, nil, "", noop, fmt.Errorf("file session store: %w", err)
	}
	return store, sessStore, nil, "file-backed " + cpath, noop, nil
}

func isInteractiveStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func main() {
	rawArgs := os.Args[1:]
	if len(rawArgs) > 0 {
		switch rawArgs[0] {
		case "version", "--version", "-v":
			fmt.Println("Mímir", version)
			return
		case "-h", "--help", "help":
			fmt.Println(usage)
			return
		}
	}

	// Parse flags across any position
	addr := ":8420"
	webFlag := false
	tuiFlag := false
	statusFlag := false
	permMode := agent.ModeEdit
	effort := ""
	modelOverride := ""
	promptFlag := ""
	var positional []string

	for i := 0; i < len(rawArgs); i++ {
		arg := rawArgs[i]
		switch {
		case arg == "--web":
			webFlag = true
		case arg == "tui" || arg == "-i" || arg == "--tui":
			tuiFlag = true
		case arg == "status" || arg == "--status":
			statusFlag = true
		case (arg == "--addr" || arg == "-addr") && i+1 < len(rawArgs):
			addr = rawArgs[i+1]
			i++
		case strings.HasPrefix(arg, "--addr="):
			addr = strings.TrimPrefix(arg, "--addr=")
		case (arg == "-p" || arg == "--prompt") && i+1 < len(rawArgs):
			promptFlag = rawArgs[i+1]
			i++
		case strings.HasPrefix(arg, "--prompt="):
			promptFlag = strings.TrimPrefix(arg, "--prompt=")
		case arg == "--mode" && i+1 < len(rawArgs):
			permMode = rawArgs[i+1]
			i++
		case strings.HasPrefix(arg, "--mode="):
			permMode = strings.TrimPrefix(arg, "--mode=")
		case arg == "--effort" && i+1 < len(rawArgs):
			effort = rawArgs[i+1]
			i++
		case strings.HasPrefix(arg, "--effort="):
			effort = strings.TrimPrefix(arg, "--effort=")
		case arg == "--model" && i+1 < len(rawArgs):
			modelOverride = rawArgs[i+1]
			i++
		case strings.HasPrefix(arg, "--model="):
			modelOverride = strings.TrimPrefix(arg, "--model=")
		default:
			positional = append(positional, arg)
		}
	}

	cwd, _ := os.Getwd()
	toolReg := tools.Default(cwd)

	home := os.Getenv("MIMIR_HOME")
	if home == "" {
		home = filepath.Join(cwd, ".mimir")
	}
	store, sessStore, authStore, storeDesc, cleanup, err := buildCortex(home)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cortex error:", err)
		os.Exit(1)
	}
	defer cleanup()
	if t, ok := toolReg.Get("ingest_repo"); ok {
		if it, ok := t.(*tools.IngestRepoTool); ok {
			it.Store = store
		}
	}

	var provider llm.Provider
	model := ""
	local := false
	if ep, ok := config.AutoDiscoverProvider(context.Background()); ok {
		if modelOverride != "" {
			ep.Model = modelOverride
		}
		provider = &llm.OpenAIProvider{
			IDStr: ep.ID, BaseURL: ep.BaseURL, APIKey: ep.APIKey, Model: ep.Model,
		}
		model = ep.Model
		local = isLocalURL(ep.BaseURL)
	}
	tracker := llm.NewTracker()

	a := agent.New(agent.Config{
		Provider: provider,
		Tools:    toolReg,
		Cortex:   store,
		Model:    model,
		Usage:    tracker,
		Local:    local,
		Mode:     permMode,
		Effort:   effort,
	})

	if len(positional) > 0 && (positional[0] == "setup" || positional[0] == "init") {
		fmt.Printf("Mímir %s · Bootstrapping System-One Decision Engine (%s)...\n", version, decision.PrimaryDecisionModel)
		dm, err := decision.EnsureLocalDecisionModel(context.Background(), "", true)
		if err != nil {
			fmt.Printf("  note: local Ollama pull skipped (%v); zero-latency heuristic engine active.\n", err)
		} else {
			fmt.Printf("  ✓ Local System-One Decision Model ready: %s\n", dm)
		}
		return
	}

	if len(positional) > 0 && (positional[0] == "ingest" || positional[0] == "gitingest") {
		src := "."
		if len(positional) > 1 {
			src = positional[1]
		}
		outFile := filepath.Join(home, "digest.txt")
		res, err := ingest.Run(context.Background(), ingest.Options{
			Source:        src,
			WriteDigestTo: outFile,
		}, store)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ingest error:", err)
			os.Exit(1)
		}
		fmt.Println(res.Summary)
		fmt.Printf("Digest written to: %s\n", outFile)
		fmt.Printf("Yggdrasil Mind Map saved to Cortex (%d neurons)\n", res.NeuronsSaved)
		return
	}

	if len(positional) > 0 && (positional[0] == "providers" || positional[0] == "subscribe") {
		fmt.Println("Mímir Well Subscriptions & BYOK Provider Catalog:")
		for _, sub := range config.BuiltinSubscriptionTiers() {
			fmt.Printf("  * %s (%s) - %s\n", sub.Name, sub.Price, sub.Description)
		}
		fmt.Println("\nBuilt-in Providers:")
		for _, p := range config.BuiltinProviders() {
			st := "unset"
			if p.Configured {
				st = "ready"
			}
			fmt.Printf("  [%s] %-28s default=%-22s (%s)\n", p.BillingMode, p.Name, p.DefaultModel, st)
		}
		return
	}

	if len(positional) > 0 && (positional[0] == "sdlc" || positional[0] == "phase") {
		mgr := a.SDLC()
		if len(positional) > 1 {
			sub := strings.ToLower(strings.TrimSpace(positional[1]))
			switch sub {
			case "agile", "waterfall", "devops":
				if err := mgr.SetMethodology(sub); err != nil {
					fmt.Fprintln(os.Stderr, "sdlc:", err)
					os.Exit(1)
				}
			case "next", "advance":
				if _, err := mgr.Advance(); err != nil {
					fmt.Fprintln(os.Stderr, "sdlc:", err)
					os.Exit(1)
				}
			default:
				if _, err := mgr.SetPhase(sub); err != nil {
					fmt.Fprintln(os.Stderr, "sdlc:", err)
					os.Exit(1)
				}
			}
		}
		fmt.Println(mgr.FormatBoard())
		return
	}

	if len(positional) > 0 && (positional[0] == "audit" || positional[0] == "critic") {
		if len(positional) < 2 {
			fmt.Fprintln(os.Stderr, "usage: mimir audit <draft-file> [retrieved-context-file]")
			os.Exit(2)
		}
		draftBytes, err := os.ReadFile(positional[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit read draft:", err)
			os.Exit(1)
		}
		var ctxBuilder strings.Builder
		if len(positional) > 2 {
			if ctxBytes, err := os.ReadFile(positional[2]); err == nil {
				ctxBuilder.Write(ctxBytes)
			}
		} else {
			for _, n := range store.All() {
				ctxBuilder.WriteString(n.Title + "\n" + n.Content + "\n")
			}
		}
		rep := critic.AuditDeterministic(ctxBuilder.String(), string(draftBytes))
		fmt.Println(rep.FormatMarkdown())
		if !rep.PassAudit {
			os.Exit(1)
		}
		return
	}

	if len(positional) > 0 && positional[0] == "engines" {
		sub := ""
		if len(positional) > 1 {
			sub = positional[1]
		}
		switch sub {
		case "", "status", "list":
			fmt.Println(engines.FormatStatus(cwd))
		case "publish":
			m, err := engines.BuildManifest(context.Background(), cwd)
			if err != nil {
				fmt.Fprintln(os.Stderr, "engines publish:", err)
				os.Exit(1)
			}
			for _, e := range m.Engines {
				fmt.Printf("  %-16s v%-8s sha256=%s...\n", e.ID, e.Version, e.SHA256[:12])
			}
			fmt.Println("wrote engines/bin/manifest.json; commit and push main to release.")
		case "run":
			if len(positional) < 3 {
				fmt.Fprintln(os.Stderr, "usage: mimir engines run <engine-id> [args...]")
				os.Exit(2)
			}
			out, err := engines.Invoke(context.Background(), cwd, positional[2], positional[3:]...)
			fmt.Println(out)
			if err != nil {
				fmt.Fprintln(os.Stderr, "engines run:", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unknown engines subcommand %q (status, run, publish)\n", sub)
			os.Exit(2)
		}
		return
	}

	if len(positional) > 0 && positional[0] == "update" {
		checkOnly := len(positional) > 1 && (positional[1] == "--check" || positional[1] == "check")
		ctx := context.Background()
		base := engines.UpdateBase()
		if !checkOnly {
			hr, err := engines.UpdateHarness(ctx, cwd)
			fmt.Printf("harness (%s): %s\n", hr.Mode, hr.Output)
			if err != nil {
				fmt.Fprintln(os.Stderr, "harness update:", err)
				os.Exit(1)
			}
			if hr.Rebuilt {
				fmt.Println("harness rebuilt: bin/mimir updated (restart to use it)")
			}
		}
		plan, err := engines.CheckForUpdates(ctx, cwd, base)
		if err != nil {
			fmt.Fprintln(os.Stderr, "engine update check:", err)
			os.Exit(1)
		}
		if len(plan.Changes) == 0 {
			fmt.Println("engines: all up to date with", base)
			return
		}
		for _, c := range plan.Changes {
			fmt.Printf("  %-16s %s -> %s (%s)\n", c.ID, c.LocalVersion, c.RemoteVersion, c.Reason)
		}
		if checkOnly {
			fmt.Println("run `mimir update` to install.")
			return
		}
		updated, err := engines.ApplyUpdates(ctx, cwd, plan)
		if err != nil {
			fmt.Fprintln(os.Stderr, "engine update:", err)
			os.Exit(1)
		}
		fmt.Println("engines updated:", strings.Join(updated, ", "))
		return
	}

	if len(positional) > 0 && (positional[0] == "audit-web" || positional[0] == "web-audit") {
		targetDir := cwd
		if len(positional) > 1 {
			targetDir = positional[1]
		}
		rep, err := tools.AuditWebDirectory(targetDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "audit-web error:", err)
			os.Exit(1)
		}
		fmt.Println(rep.FormatMarkdown())
		if !rep.PassAudit {
			os.Exit(1)
		}
		return
	}

	if len(positional) > 0 && (positional[0] == "review" || positional[0] == "ship") {
		eng, err := runengine.New(cwd, nil, runengine.DefaultLimits())
		if err != nil {
			fmt.Fprintln(os.Stderr, "runengine init:", err)
			os.Exit(1)
		}
		eng.SDLC = a.SDLC()
		if t, ok := toolReg.Get("todo"); ok {
			if tt, ok := t.(*tools.TodoTool); ok && tt.Store != nil {
				eng.Todos = tt.Store
			}
		}
		var gateOpts runengine.GateOptions
		if len(positional) > 1 {
			if draftBytes, err := os.ReadFile(positional[1]); err == nil {
				gateOpts.DraftReport = string(draftBytes)
			}
		}
		if len(positional) > 2 {
			if ctxBytes, err := os.ReadFile(positional[2]); err == nil {
				gateOpts.RetrievedContext = string(ctxBytes)
			}
		} else if gateOpts.DraftReport != "" {
			var cb strings.Builder
			for _, n := range store.All() {
				cb.WriteString(n.Title + "\n" + n.Content + "\n")
			}
			gateOpts.RetrievedContext = cb.String()
		}
		var rep runengine.GateReport
		if positional[0] == "ship" {
			rep = eng.ShipGate(context.Background(), gateOpts)
		} else {
			rep = eng.ReviewGate(context.Background(), gateOpts)
		}
		fmt.Println(rep.FormatMarkdown())
		if !rep.Passed {
			os.Exit(1)
		}
		return
	}

	if len(positional) > 0 && positional[0] == "handoff" {
		var todos []string
		if t, ok := toolReg.Get("todo"); ok {
			if tt, ok := t.(*tools.TodoTool); ok && tt.Store != nil {
				for _, item := range tt.Store.List() {
					todos = append(todos, fmt.Sprintf("- #%d [%s] %s", item.ID, item.Status, item.Content))
				}
			}
		}
		var history []llm.Message
		sid := "cli-session"
		if sessStore != nil {
			if list, err := sessStore.ListSessions(context.Background(), ""); err == nil && len(list) > 0 {
				sid = list[len(list)-1].ID
				if s, err := sessStore.GetSession(context.Background(), sid); err == nil && s != nil {
					for _, m := range s.Messages {
						history = append(history, llm.Message{Role: m.Role, Content: m.Content})
					}
				}
			}
		}
		summary := ""
		if mm, ok := ingest.LatestMindMap(store); ok {
			summary = fmt.Sprintf("%s (%d files, ~%dk tokens)", mm.RepoName, mm.TotalFiles, mm.EstTokens/1000)
		}
		md := agent.GenerateHandoffMarkdown(sid, history, todos, summary)
		outPath, err := agent.WriteHandoffFile(cwd, md)
		if err != nil {
			fmt.Fprintln(os.Stderr, "handoff error:", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Handoff generated at: %s\n\n%s\n", outPath, md)
		return
	}

	if len(positional) > 0 && positional[0] == "serve" {
		fmt.Println("Mímir serving the live UI at http://localhost" + addr + "/  (Ctrl+C to stop)")
		if err := server.Serve(server.Deps{Agent: a, Store: store, Sessions: sessStore, Auth: authStore, Usage: tracker, Addr: addr}); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
		return
	}

	mode := "chat"
	if len(positional) > 0 && (positional[0] == "chat" || positional[0] == "trace") {
		mode = positional[0]
		positional = positional[1:]
	}
	prompt := strings.TrimSpace(promptFlag)
	if prompt == "" {
		prompt = strings.TrimSpace(strings.Join(positional, " "))
	}

	if prompt != "" {
		if provider == nil {
			fmt.Fprintln(os.Stderr, "no provider configured: run Ollama locally (:11434) or set MIMIR_BASE_URL (+ MIMIR_API_KEY, MIMIR_MODEL)")
			os.Exit(2)
		}
		if mode == "trace" {
			res, err := a.RunFull(context.Background(), prompt, nil)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			for _, c := range res.Recalled {
				fmt.Printf("  [memory] %s\n", strings.Join(strings.Fields(c), " "))
			}
			for i, s := range res.Trace {
				fmt.Printf("[step %d] tool=%s args=%s\n", i+1, s.Name, s.Args)
				fmt.Printf("          -> %s\n", strings.TrimSpace(s.Result))
				if s.Err != "" {
					fmt.Printf("          !! %s\n", s.Err)
				}
			}
			fmt.Println("---")
			fmt.Println(res.Reply)
			return
		}
		reply, err := a.Run(context.Background(), prompt, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(reply)
		return
	}

	// Start background Web UI if --web was passed, or expose a callback for /web in TUI.
	startWeb := func() (string, error) {
		url := "http://localhost" + addr + "/app"
		go func() {
			_ = server.Serve(server.Deps{Agent: a, Store: store, Sessions: sessStore, Auth: authStore, Usage: tracker, Addr: addr})
		}()
		return url, nil
	}
	webURL := ""
	if webFlag {
		webURL, _ = startWeb()
	}

	// Launch interactive TUI when invoked as `mimir` in an interactive terminal, `mimir tui`, or `mimir --web`.
	if !statusFlag && (tuiFlag || webFlag || isInteractiveStdin()) {
		if err := tui.Run(context.Background(), tui.Options{
			Version:    version,
			Agent:      a,
			Tools:      toolReg,
			Cortex:     store,
			Sessions:   sessStore,
			Usage:      tracker,
			StoreDesc:  storeDesc,
			WebURL:     webURL,
			StartWebFn: startWeb,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "tui error:", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Mímir %s - the agent that remembers\n", version)
	names := make([]string, 0, len(toolReg.All()))
	for _, t := range toolReg.All() {
		names = append(names, t.Name())
	}
	fmt.Println("tools:", strings.Join(names, ", "))
	fmt.Println("memory:", storeDesc)
	if provider == nil {
		fmt.Println("provider: (none configured)")
		fmt.Println("set MIMIR_BASE_URL + MIMIR_API_KEY (+ MIMIR_MODEL), then: mimir \"your prompt\"  or  mimir trace \"...\"")
		return
	}
	fmt.Println("provider:", provider.ID(), "· model:", model)
}
