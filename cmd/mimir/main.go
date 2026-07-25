// Command mimir is the entrypoint for Mímir - the agent that remembers.
package main

import (
	"fmt"

	"github.com/David2024patton/Mimir/internal/agent"
	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

const version = "0.1.0-dev"

func main() {
	fmt.Printf("Mímir %s - the agent that remembers\n", version)

	// Walking skeleton: wire an agent with an in-memory Cortex, an empty tool
	// registry, and no provider yet. The full wiring (real providers, Docker,
	// the SDLC loop, the GUI) builds on this shape.
	store := cortex.NewMemoryStore()
	_ = agent.New(agent.Config{
		Provider: (llm.Provider)(nil), // wired to a real provider next
		Tools:    tools.NewRegistry(),
		Cortex:   store,
		Model:    "ollama/qwen3.6-35b-a3b",
	})

	fmt.Println("walking skeleton wired: agent + Cortex + tools ready")
	fmt.Println("next: wire a real provider, then the SDLC loop, then the GUI")
}
