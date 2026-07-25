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
  mimir                       show status (provider wired or not)
  mimir "your prompt"         run one turn against the configured provider
  mimir chat "your prompt"    same as above
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
		case "chat":
			args = args[1:]
		}
	}

	store := cortex.NewMemoryStore()
	toolReg := tools.NewRegistry()

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

	if prompt := strings.TrimSpace(strings.Join(args, " ")); prompt != "" {
		if provider == nil {
			fmt.Fprintln(os.Stderr, "no provider configured: set MIMIR_BASE_URL (+ MIMIR_API_KEY, MIMIR_MODEL) to run a prompt")
			os.Exit(2)
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
	if provider == nil {
		fmt.Println("wired: agent + Cortex + tools ready (no provider configured)")
		fmt.Println("set MIMIR_BASE_URL + MIMIR_API_KEY (+ MIMIR_MODEL), then: mimir \"your prompt\"")
		return
	}
	fmt.Println("provider:", provider.ID(), "· model:", model)
	fmt.Println("run: mimir \"your prompt\"")
}
