package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/David2024patton/Mimir/internal/cortex"
	"github.com/David2024patton/Mimir/internal/llm"
	"github.com/David2024patton/Mimir/internal/tools"
)

// TestAgentRunRecallAndRemember proves the full loop against a fake provider:
// a seeded memory is recalled into the system prompt, the model reply is returned,
// and the exchange is written back to the Cortex as a new neuron.
func TestAgentRunRecallAndRemember(t *testing.T) {
	var captured []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		captured = req.Messages
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "the answer"}},
			},
		})
	}))
	defer srv.Close()

	store := cortex.NewMemoryStore()
	if _, err := store.PutNeuron(context.Background(), cortex.Neuron{
		Kind: cortex.KindMemory, Layer: "preference",
		Content: "note about format my code please: use tabs", Decay: 1,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	a := New(Config{
		Provider: &llm.OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"},
		Tools:    tools.NewRegistry(),
		Cortex:   store,
		Model:    "m",
	})

	reply, err := a.Run(context.Background(), "format my code please")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if reply != "the answer" {
		t.Errorf("reply = %q", reply)
	}
	if len(captured) == 0 || captured[0].Role != "system" || !strings.Contains(captured[0].Content, "use tabs") {
		t.Errorf("expected recalled memory in system prompt, got %+v", captured)
	}

	remembered, _ := store.Search(context.Background(), "the answer", 5)
	found := false
	for _, n := range remembered {
		if strings.Contains(n.Content, "the answer") {
			found = true
		}
	}
	if !found {
		t.Error("expected the exchange to be remembered in the Cortex")
	}
}
