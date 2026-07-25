// Command mimir is the entrypoint for Mímir - the agent that remembers.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/config"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

const version = "0.1.0-dev"

const usage = `Mímir - the agent that remembers

Usage:
  mimir                       show status (provider + tools wired or not)
  mimir "your prompt"         run one turn (final reply only)
  mimir chat "your prompt"    same as above
  mimir trace "your prompt"   run one turn and print every tool call (observe the agent act)
  mimir version               print version

Provider config (environment):
  MIMIR_BASE_URL    OpenAI-compatible base URL, e.g. https://api.openai.com/v1
  MIMIR_API_KEY     bearer key (optional for local servers)
  MIMIR_MODEL       model id (optional)
  MIMIR_PROVIDER_ID provider id label (default: openai-compatible)`

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
	store := cortex.NewMemoryStore()
	toolReg := tools.Default(cwd)

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
			reply, trace, err := a.RunTrace(context.Background(), prompt)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			for i, s := range trace {
				fmt.Printf("[step %d] tool=%s args=%s\n", i+1, s.Name, s.Args)
				fmt.Printf("          -> %s\n", strings.TrimSpace(s.Result))
				if s.Err != "" {
					fmt.Printf("          !! %s\n", s.Err)
				}
			}
			fmt.Println("---")
			fmt.Println(reply)
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
	if provider == nil {
		fmt.Println("provider: (none configured)")
		fmt.Println("set MIMIR_BASE_URL + MIMIR_API_KEY (+ MIMIR_MODEL), then: mimir \"your prompt\"  or  mimir trace \"...\"")
		return
	}
	fmt.Println("provider:", provider.ID(), "· model:", model)
}
