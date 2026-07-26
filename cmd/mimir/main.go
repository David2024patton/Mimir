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
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/server"
	"github.com/David2024patton/Mimir/internal/tools"
)

const version = "0.1.0-dev"

const usage = `Mímir - the agent that remembers

Usage:
  mimir                       show status (provider + tools + memory file)
  mimir "your prompt"         run one turn (final reply only)
  mimir chat "your prompt"    same as above
  mimir trace "your prompt"   run one turn and print recalled memory + every tool call
  mimir serve [--addr :8420]  serve the live book-spine UI + HTTP wire (E70)
  mimir version               print version

Provider config (environment):
  MIMIR_BASE_URL    OpenAI-compatible base URL, e.g. https://api.openai.com/v1
  MIMIR_API_KEY     bearer key (optional for local servers)
  MIMIR_MODEL       model id (optional)
  MIMIR_PROVIDER_ID provider id label (default: openai-compatible)

Memory (the Cortex) persists across runs to a JSON file:
  MIMIR_HOME        memory dir (default: ./.mimir in the working directory)
  MIMIR_CORTEX      full path to the cortex file (default: <MIMIR_HOME>/cortex.json)

Cortex backend (E6): set MIMIR_CORTEX_BACKEND=surreal for the SurrealDB brain
(managed sidecar + hybrid vector/full-text recall). Related env:
  MIMIR_CORTEX_BACKEND  file (default) | surreal
  MIMIR_SURREAL_ADDR    sidecar bind address (default 127.0.0.1:8000)
  MIMIR_SURREAL_DATA    sidecar data dir (default <MIMIR_HOME>/surreal; "memory" = in-memory)
  MIMIR_SURREAL_BIN     path to the surreal binary (auto-located if unset)
  MIMIR_EMBED_URL       embeddings server (default: Ollama at MIMIR_BASE_URL or :11434)
  MIMIR_EMBED_MODEL     embedding model (default nomic-embed-text)
  MIMIR_EMBED_DIM       embedding dimension (default 768)`

// isLocalURL reports whether a provider base URL points at a local server (Ollama,
// vLLM, Mímir's own gateway). Local tokens cost $0 on the dashboard (F52).
func isLocalURL(u string) bool {
	return strings.Contains(u, "localhost") || strings.Contains(u, "127.0.0.1") || strings.Contains(u, "0.0.0.0")
}

// buildCortex creates the Cortex store for the configured backend. MIMIR_CORTEX_BACKEND
// selects it: "surreal" starts a managed SurrealDB sidecar with an Ollama embedder for
// hybrid vector + full-text recall (E6); anything else (default) uses the file-backed
// store. The returned cleanup stops the sidecar, if one was started. For SurrealDB,
// it also returns a SessionStore for F2.2 conversation persistence.
func buildCortex(home string) (cortex.Store, cortex.SessionStore, *cortex.AuthStore, string, func(), error) {
	noop := func() {}
	if strings.EqualFold(os.Getenv("MIMIR_CORTEX_BACKEND"), "surreal") {
		addr := os.Getenv("MIMIR_SURREAL_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8000"
		}
		dataPath := os.Getenv("MIMIR_SURREAL_DATA")
		if dataPath == "" {
			dataPath = filepath.Join(home, "surreal")
		}
		sc := &cortex.Sidecar{Addr: addr, DataPath: dataPath}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if err := sc.Start(ctx); err != nil {
			return nil, nil, nil, "", noop, fmt.Errorf("surreal sidecar: %w", err)
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
		store, err := cortex.NewSurrealStore(context.Background(), cortex.SurrealConfig{
			Addr:      sc.HTTPAddr(),
			Dimension: dim,
			Embedder:  &cortex.OllamaEmbedder{BaseURL: embedURL, Model: embedModel},
		})
		if err != nil {
			sc.Stop()
			return nil, nil, nil, "", noop, fmt.Errorf("surreal store: %w", err)
		}
		cli := store.Client()
		// Create session store for F2.2 conversation persistence
		sessStore := cortex.NewSessionStore(cli)
		if err := sessStore.EnsureSessionsTable(context.Background()); err != nil {
			sc.Stop()
			return nil, nil, nil, "", noop, fmt.Errorf("sessions table: %w", err)
		}
		// Create auth store
		authStore := cortex.NewAuthStore(cli)
		if err := authStore.EnsureAuthTables(context.Background()); err != nil {
			sc.Stop()
			return nil, nil, nil, "", noop, fmt.Errorf("auth tables: %w", err)
		}
		// Seed super admin
		adminEmail := os.Getenv("MIMIR_ADMIN_EMAIL")
		if adminEmail == "" {
			adminEmail = "david@itak.live"
		}
		if err := authStore.SeedAdmin(context.Background(), adminEmail); err != nil {
			sc.Stop()
			return nil, nil, nil, "", noop, fmt.Errorf("seed admin: %w", err)
		}
		fmt.Fprintf(os.Stderr, "auth: super admin %s seeded\n", adminEmail)
		desc := fmt.Sprintf("SurrealDB %s (embed %s, dim %d)", sc.HTTPAddr(), embedModel, dim)
		return store, sessStore, authStore, desc, sc.Stop, nil
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
	return store, sessStore, nil, "file-backed "+cpath, noop, nil
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "version", "--version", "-v":
			fmt.Println("Mímir", version)
			return
		case "-h", "--help", "help":
			fmt.Println(usage)
			return
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

	var provider llm.Provider
	model := ""
	local := false
	if ep, ok := config.ProviderFromEnv(); ok {
		provider = &llm.OpenAIProvider{
			IDStr: ep.ID, BaseURL: ep.BaseURL, APIKey: ep.APIKey, Model: ep.Model,
		}
		model = ep.Model
		local = isLocalURL(ep.BaseURL)
	}
	tracker := llm.NewTracker()

	a := agent.New(agent.Config{
		Provider: provider, Tools: toolReg, Cortex: store, Model: model,
		Usage: tracker, Local: local,
	})

	if len(args) > 0 && args[0] == "serve" {
		addr := ":8420"
		for i := 1; i < len(args); i++ {
			if (args[i] == "--addr" || args[i] == "-addr") && i+1 < len(args) {
				addr = args[i+1]
				i++
			} else if strings.HasPrefix(args[i], "--addr=") {
				addr = strings.TrimPrefix(args[i], "--addr=")
			}
		}
		fmt.Println("Mímir serving the live UI at http://localhost" + addr + "/  (Ctrl+C to stop)")
		if err := server.Serve(server.Deps{Agent: a, Store: store, Sessions: sessStore, Auth: authStore, Usage: tracker, Addr: addr}); err != nil {
			fmt.Fprintln(os.Stderr, "serve:", err)
			os.Exit(1)
		}
		return
	}

	mode := "chat"
	if len(args) > 0 && (args[0] == "chat" || args[0] == "trace") {
		mode = args[0]
		args = args[1:]
	}
	prompt := strings.TrimSpace(strings.Join(args, " "))

	if prompt != "" {
		if provider == nil {
			fmt.Fprintln(os.Stderr, "no provider configured: set MIMIR_BASE_URL (+ MIMIR_API_KEY, MIMIR_MODEL) to run a prompt")
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
