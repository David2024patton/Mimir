// Command mimir is the entrypoint for Mímir - the agent that remembers.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
  MIMIR_CORTEX      full path to the cortex file (default: <MIMIR_HOME>/cortex.json)`

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
	cpath := os.Getenv("MIMIR_CORTEX")
	if cpath == "" {
		cpath = filepath.Join(home, "cortex.json")
	}
	store, err := cortex.NewMemoryStoreAt(cpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cortex error:", err)
		os.Exit(1)
	}

	var provider llm.Provider
	model := ""
	if ep, ok := config.ProviderFromEnv(); ok {
		provider = &llm.OpenAIProvider{
			IDStr: ep.ID, BaseURL: ep.BaseURL, APIKey: ep.APIKey, Model: ep.Model,
		}
		model = ep.Model
	}

	a := agent.New(agent.Config{
		Provider: provider, Tools: toolReg, Cortex: store, Model: model,
	})

	if len(args) > 0 && args[0] == "serve" {
		addr := ":8420"
		for i := 1; i < len(args); i++ {
			if args[i] == "--addr" && i+1 < len(args) {
				addr = args[i+1]
				i++
			}
		}
		fmt.Println("Mímir serving the live UI at http://localhost" + addr + "/  (Ctrl+C to stop)")
		if err := server.Serve(server.Deps{Agent: a, Store: store, Addr: addr}); err != nil {
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
			res, err := a.RunFull(context.Background(), prompt)
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
		reply, err := a.Run(context.Background(), prompt)
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
	fmt.Println("memory:", cpath)
	if provider == nil {
		fmt.Println("provider: (none configured)")
		fmt.Println("set MIMIR_BASE_URL + MIMIR_API_KEY (+ MIMIR_MODEL), then: mimir \"your prompt\"  or  mimir trace \"...\"")
		return
	}
	fmt.Println("provider:", provider.ID(), "· model:", model)
}
